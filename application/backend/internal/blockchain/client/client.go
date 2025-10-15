package client

import (
	"fiatless/internal/ijson"
)

type BlockchainClient interface {
	ExecuteCommand(command string, params any) (map[string]any, error)
}

type IJSONBlockchainClient struct {
	ijsonClient *ijson.IJSONClient
}

func NewIJSONBlockchainClient(ijsonClient *ijson.IJSONClient) *IJSONBlockchainClient {
	return &IJSONBlockchainClient{
		ijsonClient: ijsonClient,
	}
}

func (c *IJSONBlockchainClient) ExecuteCommand(command string, params any) (map[string]any, error) {
	return c.ijsonClient.InvokeCommand(command, params)
}
