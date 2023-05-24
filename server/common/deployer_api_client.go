package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type DeployerApiClient struct {
	url string
}

func NewDeployerApiClient(url string) DeployerApiClient {
	return DeployerApiClient{
		url: url,
	}
}

func (d *DeployerApiClient) DeployEnclave(name string) (string, error) {
	deployerURL := fmt.Sprintf("%s/enclaves", d.url)

	type payload struct {
		Name string `json:"name"`
	}

	body, err := json.Marshal(payload{name})
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, deployerURL, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to instantiate request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return "", fmt.Errorf("received error status code: %d", res.StatusCode)
	}

	type input struct {
		EnclaveURL string `json:"url"`
	}

	var in input
	if err := json.NewDecoder(res.Body).Decode(&in); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return in.EnclaveURL, nil
}

func (d *DeployerApiClient) RemoveEnclave(name string) error {
	deployerURL := fmt.Sprintf("%s/enclaves/%s", d.url, name)

	req, err := http.NewRequest(http.MethodDelete, deployerURL, nil)
	if err != nil {
		return fmt.Errorf("failed to instantiate request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return fmt.Errorf("received error status code: %d", res.StatusCode)
	}

	return nil
}
