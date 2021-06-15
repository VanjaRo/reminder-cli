package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type reminderBody struct {
	ID       string        `json:"id"`
	Title    string        `json:"title"`
	Message  string        `json:"message"`
	Duration time.Duration `json:"duration"`
}

type HTTPClient struct {
	client     *http.Client
	backendURI string
}

func NewHTTPClient(uri string) HTTPClient {
	return HTTPClient{
		backendURI: uri,
		client:     &http.Client{},
	}
}

func (c HTTPClient) apiCall(method, path string, body interface{}, resCode int) ([]byte, error) {
	bs, err := json.Marshal(body)
	if err != nil {
		return nil, wrapError("could not marshal request body", err)
	}
	req, err := http.NewRequest(method, c.backendURI+path, bytes.NewReader(bs))
	if err != nil {
		return []byte{}, wrapError("could not create request", err)
	}
	res, err := c.client.Do(req)
	if err != nil {
		return []byte{}, wrapError("could not make hhtp call", err)
	}
	resBody, err := c.readResBody(res.Body)
	if err != nil {
		return []byte{}, err
	}
	if res.StatusCode != resCode {
		if len(resBody) > 0 {
			fmt.Printf("got this response body:\n%s\n", resBody)
		}
		return []byte{}, fmt.Errorf(
			"expected response body: %d, got: %d",
			resCode, res.StatusCode,
		)
	}
	return []byte(resBody), err
}

func (c HTTPClient) readResBody(b io.Reader) (string, error) {
	bs, err := ioutil.ReadAll(b)
	if err != nil {
		return "", wrapError("could not read response body", err)
	}
	if len(bs) == 0 {
		return "", nil
	}

	var buff bytes.Buffer
	if err := json.Indent(&buff, bs, "", "\t"); err != nil {
		return "", wrapError("could not indent json", err)
	}
	return buff.String(), nil
}

func (c HTTPClient) Create(title, msg string, duration time.Duration) ([]byte, error) {
	reqBody := reminderBody{
		Title:    title,
		Message:  msg,
		Duration: duration,
	}
	return c.apiCall(http.MethodPost, "/reminders", &reqBody, http.StatusOK)
}

func (c HTTPClient) Edit(id, title, msg string, duration time.Duration) ([]byte, error) {
	reqBody := reminderBody{
		ID:       id,
		Title:    title,
		Message:  msg,
		Duration: duration,
	}
	return c.apiCall(http.MethodPatch, "/reminders/"+id, &reqBody, http.StatusOK)
}

func (c HTTPClient) Fetch(ids []string) ([]byte, error) {
	idSet := strings.Join(ids, ",")
	return c.apiCall(http.MethodGet, "/reminders/"+idSet, nil, http.StatusOK)
}

func (c HTTPClient) Delete(ids []string) error {
	idSet := strings.Join(ids, ",")
	_, err := c.apiCall(http.MethodDelete, "/reminders/"+idSet, nil, http.StatusNoContent)
	return err
}

func (c HTTPClient) Healthy(host string) bool {
	res, err := http.Get(host + "/health")
	if err != nil || res.StatusCode != http.StatusOK {
		return false
	}
	return true
}
