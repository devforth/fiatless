package utils

import (
	"context"
	api_tronpb "fiatless/pkg/proto/tron/api"
	core_tronpb "fiatless/pkg/proto/tron/core"
	"fmt"
)

type ChainParams struct {
	client      api_tronpb.WalletClient
	chainParams *core_tronpb.ChainParameters
}

func NewChainParams(client api_tronpb.WalletClient) *ChainParams {
	return &ChainParams{
		client: client,
	}
}

func (c *ChainParams) InitChainParams(ctx context.Context) error {
	params, err := c.client.GetChainParameters(ctx, &api_tronpb.EmptyMessage{})
	if err != nil {
		return err
	}
	c.chainParams = params
	return nil
}

func (c *ChainParams) GetChainParameter(ctx context.Context, key string) (int64, error) {
	if c.chainParams == nil {
		return 0, fmt.Errorf("chain parameters not initialized")
	}

	for _, param := range c.chainParams.ChainParameter {
		if param.Key == key {
			return param.Value, nil
		}
	}

	return 0, fmt.Errorf("chain parameter not found")
}
