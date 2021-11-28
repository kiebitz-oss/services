package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/crypto"
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
}

func (r *Response) JSON() (map[string]interface{}, error) {
	var value map[string]interface{}

	defer r.Response.Body.Close()

	if body, err := io.ReadAll(r.Response.Body); err != nil {
		return nil, err
	} else if err = json.Unmarshal(body, &value); err != nil {
		return nil, err
	}

	return value, nil
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

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	if resp, err := client.Do(req); err != nil {
		return nil, err
	} else {
		return &Response{Response: resp}, nil
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

func (a *AppointmentsClient) ConfirmProvider(params *services.ConfirmProviderParams, mediator *crypto.Actor) (*Response, error) {
	return nil, nil
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
	return nil, nil
}

func (a *AppointmentsClient) GetProviderAppointments(params *services.GetProviderAppointmentsParams) (*Response, error) {
	return nil, nil
}

func (a *AppointmentsClient) PublishAppointments(params interface{}) (*Response, error) {
	return nil, nil
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

func (a *AppointmentsClient) StoreProviderData(params interface{}) (*Response, error) {
	return nil, nil
}

func (a *AppointmentsClient) CheckProviderData(params interface{}) (*Response, error) {
	return nil, nil
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
