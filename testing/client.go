package testing

import (
	"bytes"
	"encoding/json"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/crypto"
	"io"
	"net/http"
)

type Client struct {
	settings     *services.Settings
	key          *crypto.Key
	client       *http.Client
	Storage      Requester
	Appointments Requester
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

func MakeClient(settings *services.Settings, key *crypto.Key) *Client {
	client := &http.Client{}
	return &Client{
		key:          key,
		settings:     settings,
		client:       client,
		Storage:      MakeAPIClient(settings.Admin.Client.StorageEndpoint, key, client),
		Appointments: MakeAPIClient(settings.Admin.Client.AppointmentsEndpoint, key, client),
	}
}

type Requester func(method string, params interface{}) (*Response, error)

func MakeAPIClient(url string, key *crypto.Key, client *http.Client) Requester {
	return func(method string, params interface{}) (*Response, error) {
		return Request(url, method, params, key, client)
	}
}

func Request(url, method string, params interface{}, key *crypto.Key, client *http.Client) (*Response, error) {

	if params == nil {
		params = map[string]interface{}{}
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
