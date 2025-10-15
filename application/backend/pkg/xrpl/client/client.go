package client

import (
	"net/http"

	"fiatless/pkg/httpclient"

	xrpl_rpc "github.com/Peersyst/xrpl-go/xrpl/rpc"
)

type XRPLClient struct {
	xrpl_rpc.Client
	httpClient *http.Client
	config     *xrpl_rpc.Config
}

func NewXRPLClient(cfg *xrpl_rpc.Config, rps float64) *XRPLClient {
	return &XRPLClient{
		Client:     *xrpl_rpc.NewClient(cfg),
		httpClient: &http.Client{Transport: httpclient.NewRateLimitedTransport(rps, nil)},
		config:     cfg,
	}
}

func (c *XRPLClient) Request(reqParams xrpl_rpc.XRPLRequest) (xrpl_rpc.XRPLResponse, error) {
	// Create a new config with our rate-limited HTTP client
	modifiedConfig := &xrpl_rpc.Config{
		URL:        c.config.URL,
		Headers:    c.config.Headers,
		HTTPClient: c.httpClient, // Use our rate-limited client
	}

	// Create a temporary client with the modified config
	tempClient := xrpl_rpc.NewClient(modifiedConfig)

	// Use the temporary client to make the request
	return tempClient.Request(reqParams)
}
