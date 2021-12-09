// Kiebitz - Privacy-Friendly Appointment Scheduling
// Copyright (C) 2021-2021 The Kiebitz Authors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/crypto"
	"github.com/kiebitz-oss/services/jsonrpc"
	"github.com/kiprotect/go-helpers/forms"
	"io"
	"net/http"
	"time"
)

type Client struct {
	settings     *services.Settings
	client       *http.Client
	Storage      *StorageClient
	Appointments *AppointmentsClient
}

type Response struct {
	*http.Response
	body []byte
}

func (r *Response) Read() error {

	if r.body != nil {
		return nil // already read the body
	}

	defer r.Response.Body.Close()

	if body, err := io.ReadAll(r.Response.Body); err != nil {
		return err
	} else {
		r.body = body
		return nil
	}
}

func (r *Response) CoerceResult(target interface{}, form *forms.Form) error {
	if bytes, err := r.Bytes(); err != nil {
		return err
	} else {
		response := &jsonrpc.Response{}
		if err := json.Unmarshal(bytes, response); err != nil {
			return err
		}
		if response.Result == nil {
			return fmt.Errorf("no result")
		}
		if form != nil {
			if mapResult, ok := response.Result.(map[string]interface{}); !ok {
				return fmt.Errorf("expected a map result")
			} else if params, err := form.Validate(mapResult); err != nil {
				return err
			} else {
				return form.Coerce(target, params)
			}
		}
		return forms.Coerce(target, response.Result)
	}
}

func (r *Response) JSON() (map[string]interface{}, error) {
	var value map[string]interface{}

	if err := r.Read(); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(r.body, &value); err != nil {
		return nil, err
	}

	return value, nil
}

func (r *Response) Bytes() ([]byte, error) {
	if err := r.Read(); err != nil {
		return nil, err
	}
	return r.body, nil
}

func MakeClient(settings *services.Settings) *Client {
	client := &http.Client{}
	return &Client{
		settings:     settings,
		client:       client,
		Storage:      MakeStorageClient(settings, client),
		Appointments: MakeAppointmentsClient(settings, client),
	}
}

type Requester func(method string, params interface{}, key *crypto.Key) (*Response, error)

func MakeAPIClient(url string, client *http.Client) Requester {
	return func(method string, params interface{}, key *crypto.Key) (*Response, error) {
		return Request(url, method, params, key, client)
	}
}

func Request(url, method string, params interface{}, key *crypto.Key, client *http.Client) (*Response, error) {

	if params == nil {
		params = map[string]interface{}{}
	}

	if key != nil {

		bytes, err := json.Marshal(params)

		if err != nil {
			return nil, err
		}

		signedData, err := key.SignString(string(bytes))

		if err != nil {
			return nil, err
		}

		params = signedData.AsMap()

	}

	jsonrpcRequest := map[string]interface{}{
		"method":  method,
		"jsonrpc": "2.0",
		"params":  params,
	}

	jsonData, err := json.Marshal(jsonrpcRequest)

	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(jsonData)

	req, err := http.NewRequest("POST", url, reader)

	// this is important, Golang won't close requests otherwise...
	// https://github.com/golang/go/issues/28012
	req.Close = true

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	if resp, err := client.Do(req); err != nil {
		return nil, err
	} else {
		response := &Response{Response: resp}
		// we always read the body as otherwise the connection can stay open
		// and will not be freed properly...
		return response, response.Read()
	}

}

type AppointmentsClient struct {
	settings  *services.Settings
	requester Requester
}

func MakeAppointmentsClient(settings *services.Settings, client *http.Client) *AppointmentsClient {
	return &AppointmentsClient{
		settings:  settings,
		requester: MakeAPIClient(settings.Admin.Client.AppointmentsEndpoint, client),
	}
}

func (a *AppointmentsClient) GetKeys() (*Response, error) {
	return a.requester("getKeys", nil, nil)
}

func (a *AppointmentsClient) ResetDB() (*Response, error) {
	signingKey := a.settings.Admin.Signing.Key("root")

	if signingKey == nil {
		return nil, fmt.Errorf("root key missing")
	}

	data := map[string]interface{}{
		"timestamp": time.Now(),
	}

	return a.requester("resetDB", data, signingKey)

}

func (a *AppointmentsClient) AddMediatorPublicKeys(mediator *crypto.Actor) (*Response, error) {
	signingKey := a.settings.Admin.Signing.Key("root")

	if signingKey == nil {
		return nil, fmt.Errorf("root key missing")
	}

	data := map[string]interface{}{
		"signing":    mediator.SigningKey.PublicKey,
		"encryption": mediator.EncryptionKey.PublicKey,
		"timestamp":  time.Now(),
	}

	return a.requester("addMediatorPublicKeys", data, signingKey)

}

type Provider struct {
	Actor      *crypto.Actor
	DataKey    *crypto.Key
	QueueData  *services.ProviderQueueData
	PublicData *services.ProviderData
}

func (a *AppointmentsClient) ConfirmProvider(provider *Provider, mediator *crypto.Actor) (*Response, error) {

	keyData := &services.KeyData{
		Signing:    provider.Actor.SigningKey.PublicKey,
		Encryption: provider.Actor.EncryptionKey.PublicKey,
		QueueData:  provider.QueueData,
	}

	signedKeyData, err := keyData.Sign(mediator.SigningKey)

	if err != nil {
		return nil, err
	}

	providerData := []byte("test")

	ephemeralKey, err := crypto.GenerateWebKey("ephemeral-mediator", "ecdh")

	if err != nil {
		return nil, err
	}

	encryptedProviderData, err := ephemeralKey.Encrypt(providerData, provider.Actor.EncryptionKey)

	if err != nil {
		return nil, err
	}

	signedEncryptedProviderData, err := encryptedProviderData.Sign(mediator.SigningKey)

	if err != nil {
		return nil, err
	}

	signedProviderData, err := provider.PublicData.Sign(mediator.SigningKey)

	if err != nil {
		return nil, err
	}

	params := &services.ConfirmProviderParams{
		Timestamp:          time.Now(),
		PublicProviderData: signedProviderData,
		EncryptedProviderData: &services.EncryptedProviderData{
			Signature: signedEncryptedProviderData.Signature,
			PublicKey: signedEncryptedProviderData.PublicKey,
			JSON:      string(signedEncryptedProviderData.Data),
			Data:      encryptedProviderData,
		},
		SignedKeyData: signedKeyData,
	}

	return a.requester("confirmProvider", params, mediator.SigningKey)
}

func (a *AppointmentsClient) AddCodes(params *services.AddCodesParams) (*Response, error) {
	return nil, nil
}

func (a *AppointmentsClient) UploadDistances(params *services.UploadDistancesParams) (*Response, error) {
	return nil, nil
}

func (a *AppointmentsClient) GetStats(params *services.GetStatsParams) (*Response, error) {
	return nil, nil
}

func (a *AppointmentsClient) GetAppointmentsByZipCode(params *services.GetAppointmentsByZipCodeParams) (*Response, error) {
	return a.requester("getAppointmentsByZipCode", params, nil)
}

func (a *AppointmentsClient) GetProviderAppointments(params *services.GetProviderAppointmentsParams) (*Response, error) {
	return nil, nil
}

func (a *AppointmentsClient) PublishAppointments(params *services.PublishAppointmentsParams, provider *Provider) (*Response, error) {
	return a.requester("publishAppointments", params, provider.Actor.SigningKey)
}

func (a *AppointmentsClient) BookAppointment(params interface{}) (*Response, error) {
	return nil, nil
}

func (a *AppointmentsClient) CancelAppointment(params interface{}) (*Response, error) {
	return nil, nil
}

func (a *AppointmentsClient) GetToken(params interface{}) (*Response, error) {
	return nil, nil
}

type ConfirmProviderData struct {
	Data       *services.ProviderData `json:"data"`
	PublicKeys *PublicKeys            `json:"publicKeys"`
}

type PublicKeys struct {
	Signing    []byte `json:"signing"`
	Encryption []byte `json:"encryption"`
}

func (a *AppointmentsClient) StoreProviderData(provider *Provider) (*Response, error) {

	dataKey := a.settings.Appointments.Key("provider")

	if dataKey == nil {
		return nil, fmt.Errorf("provider data key missing")
	}

	var err error
	provider.DataKey, err = crypto.GenerateWebKey("ephemeral-provider", "ecdh")

	if err != nil {
		return nil, err
	}

	confirmProviderData := &ConfirmProviderData{
		Data: provider.PublicData,
		PublicKeys: &PublicKeys{
			Signing:    provider.Actor.SigningKey.PublicKey,
			Encryption: provider.Actor.EncryptionKey.PublicKey,
		},
	}

	data, err := json.Marshal(confirmProviderData)

	if err != nil {
		return nil, err
	}

	encryptedProviderData, err := provider.DataKey.Encrypt(data, dataKey)
	storeProviderDataParams := &services.StoreProviderDataParams{
		EncryptedData: encryptedProviderData,
		Code:          nil,
	}

	return a.requester("storeProviderData", storeProviderDataParams, provider.Actor.SigningKey)
}

func (a *AppointmentsClient) CheckProviderData(provider *Provider) (*Response, error) {

	params := &services.CheckProviderDataParams{
		Timestamp: time.Now(),
	}
	return a.requester("checkProviderData", params, provider.Actor.SigningKey)
}

func (a *AppointmentsClient) GetPendingProviderData(params interface{}) (*Response, error) {
	return nil, nil
}

func (a *AppointmentsClient) GetVerifiedProviderData(params interface{}) (*Response, error) {
	return nil, nil
}

type StorageClient struct {
	settings  *services.Settings
	requester Requester
}

func MakeStorageClient(settings *services.Settings, client *http.Client) *StorageClient {
	return &StorageClient{
		settings:  settings,
		requester: MakeAPIClient(settings.Admin.Client.StorageEndpoint, client),
	}
}

func (a *StorageClient) ResetDB() (*Response, error) {
	signingKey := a.settings.Admin.Signing.Key("root")

	if signingKey == nil {
		return nil, fmt.Errorf("root key missing")
	}

	data := map[string]interface{}{
		"timestamp": time.Now(),
	}

	return a.requester("resetDB", data, signingKey)

}
