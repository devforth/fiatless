package grpc

import (
	grpcclient "fiatless/pkg/grpcclient"
	api_tronpb "fiatless/pkg/proto/tron/api"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TronClient struct {
	api_tronpb.WalletClient
	interceptor *grpcclient.RateLimitedInterceptor
	cc          grpc.ClientConnInterface
}

func NewTronClient(target string, rps float64) (*TronClient, error) {
	interceptor := grpcclient.NewRateLimitedInterceptor(rps)

	cc, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(interceptor.UnaryClientInterceptor()))
	if err != nil {
		return nil, err
	}

	walletClient := api_tronpb.NewWalletClient(cc)

	return &TronClient{
		WalletClient: walletClient,
		cc:           cc,
		interceptor:  interceptor,
	}, nil
}

func (c *TronClient) UpdateRateLimit(rps float64) {
	c.interceptor.UpdateRPS(rps)
}
