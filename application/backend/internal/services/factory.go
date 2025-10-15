package services

import (
	"fiatless/internal/blockchain/client"
	"fiatless/internal/ijson"
	"fiatless/internal/services/bitcoin"
	"fiatless/internal/services/bsc"
	"fiatless/internal/services/ethereum"
	"fiatless/internal/services/solana"
	"fiatless/internal/services/tron"
	"fiatless/internal/services/xrpl"
)

type ServiceFactory struct {
	blockchainClient client.BlockchainClient
}

func NewServiceFactory(ijsonClient *ijson.IJSONClient) *ServiceFactory {
	blockchainClient := client.NewIJSONBlockchainClient(ijsonClient)
	return &ServiceFactory{
		blockchainClient: blockchainClient,
	}
}

func (f *ServiceFactory) CreateEthereumService() *ethereum.Service {
	return ethereum.NewService(f.blockchainClient)
}

func (f *ServiceFactory) CreateTronService() *tron.Service {
	return tron.NewService(f.blockchainClient)
}

func (f *ServiceFactory) CreateBSCService() *bsc.Service {
	return bsc.NewService(f.blockchainClient)
}

func (f *ServiceFactory) CreateBitcoinService() *bitcoin.Service {
	return bitcoin.NewService(f.blockchainClient)
}

func (f *ServiceFactory) CreateSolanaService() *solana.Service {
	return solana.NewService(f.blockchainClient)
}

func (f *ServiceFactory) CreateXRPLService() *xrpl.Service {
	return xrpl.NewService(f.blockchainClient)
}
