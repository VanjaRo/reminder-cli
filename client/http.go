package client

import (
	"net/http"
	"time"
)

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

func (c HTTPClient) Create(title, msg string, duration time.Duration) ([]byte, error) {
	res := []byte(`response for create reminder`)
	return res, nil
}

func (c HTTPClient) Edit(id, title, msg string, duration time.Duration) ([]byte, error) {
	res := []byte(`response for edit reminder`)
	return res, nil
}

func (c HTTPClient) Fetch(ids []string) ([]byte, error) {
	res := []byte(`response for fetch reminder`)
	return res, nil
}

func (c HTTPClient) Delete(ids []string) error {
	return nil
}

func (c HTTPClient) Healthy(host string) bool {
	return true
}
