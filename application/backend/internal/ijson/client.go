package ijson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// IJSONClient represents the interface for interacting with the ijson message broker
type IIJSONClient interface {
	// Worker methods
	GetCommand(commandPath string) (map[string]any, error)
	SendResult(result map[string]any) error

	// Client methods
	InvokeCommand(commandPath string, params any) (map[string]any, error)
	GetCommandLongPolling(path string) (map[string]any, error)
}

// DefaultIJSONClient implements the IJSONClient interface
type IJSONClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewIJSONClient creates a new ijson client
func NewIJSONClient(baseURL string) *IJSONClient {
	return &IJSONClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 0,
		},
	}
}

// GetCommand waits for a command to be available (worker side)
// This is equivalent to: curl localhost:8001/test/command -H 'type: get'
func (c *IJSONClient) GetCommand(commandPath string) (map[string]any, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, commandPath)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Type", "get")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// SendResult sends a response back to the client (worker side)
// This is equivalent to: curl localhost:8001 -H 'type: result' -d '{"id": 123, "result": "data received"}'
func (c *IJSONClient) SendResult(result map[string]any) error {
	data, err := json.Marshal(result)
	log.Printf("result: %s", string(data))
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Type", "result")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return nil
}

// InvokeCommand calls a remote command and waits for the response (client side)
// This is equivalent to: curl localhost:8001/test/command -d '{"id": 123, "params": "test data"}'
func (c *IJSONClient) InvokeCommand(commandPath string, params any) (map[string]any, error) {
	// Generate a unique request ID
	requestID := strconv.FormatInt(time.Now().UnixNano(), 10)

	request := map[string]any{
		"id":     requestID,
		"params": params,
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s%s", c.baseURL, commandPath)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *IJSONClient) GetCommandLongPolling(path string) (map[string]any, error) {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("type", "get")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
