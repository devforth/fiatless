package tron

import (
	"context"
	"fiatless/internal/blockchain/handler"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/ijson"
	"fiatless/internal/models"
	"fiatless/pkg/tron"
	"fiatless/pkg/tron/address"
	"sync"
)

type BalanceHandler struct {
	handler.TronHandler
}

func NewBalanceHandler(client *tron.Tron) *BalanceHandler {
	return &BalanceHandler{
		TronHandler: handler.NewTronHandler(client),
	}
}

func (h *BalanceHandler) CommandPath() string {
	return "/tron/wallet/balance"
}

func (h *BalanceHandler) Handle(ctx context.Context, client *ijson.IJSONClient, command map[string]any) error {
	var params bp_models.TronWalletBalanceRequest
	requestID, err := h.ParseParams(command, &params)
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}

	balance, err := h.getBalance(params)
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}

	return client.SendResult(map[string]any{
		"id":     requestID,
		"result": balance,
	})
}

func (h *BalanceHandler) getBalance(params bp_models.TronWalletBalanceRequest) (*models.TronWalletBalance, error) {
	resultBalance := models.TronWalletBalance{}
	addr := params.Address

	var wg sync.WaitGroup

	errCh := make(chan error, len(params.TRC20Tokens)+2)

	var mu sync.Mutex

	for _, token := range params.TRC20Tokens {
		wg.Add(1)
		go func(tkn address.TronAddress) {
			defer wg.Done()

			balance, err := h.Tron.GetAccountTRC20Balance(&addr, &tkn)
			if err != nil {
				errCh <- err
				return
			}

			mu.Lock()
			resultBalance.TRC20Balances = append(resultBalance.TRC20Balances, models.TRC20Balance{
				ContractAddress: tkn,
				Balance:         balance,
			})
			mu.Unlock()
		}(token)
	}

	if params.IncludeTRX {
		wg.Add(1)
		go func() {
			defer wg.Done()

			balance, err := h.Tron.GetAccountTRXBalance(&addr)
			if err != nil {
				errCh <- err
				return
			}

			mu.Lock()
			resultBalance.TRXBalance = balance
			mu.Unlock()
		}()
	}

	if params.IncludeResources {
		wg.Add(1)
		go func() {
			defer wg.Done()

			resources, err := h.Tron.GetAccountResource(&addr)
			if err != nil {
				errCh <- err
				return
			}

			mu.Lock()
			resultBalance.AvailableBandwidth = resources.FreeNetLimit + resources.NetLimit - resources.NetUsed
			resultBalance.AvailableEnergy = resources.EnergyLimit - resources.EnergyUsed
			resultBalance.TotalEnergy = resources.EnergyLimit
			resultBalance.TotalBandwidth = resources.NetLimit + resources.FreeNetLimit
			mu.Unlock()
		}()
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}

	return &resultBalance, nil
}
