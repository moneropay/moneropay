/*
 * MoneroPay is a Monero payment processor.
 * Copyright (C) 2022 Laurynas Četyrkinas <stnby@kernal.eu>
 * Copyright (C) 2022 İrem Kuyucu <siren@kernal.eu>
 *
 * MoneroPay is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * MoneroPay is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with MoneroPay.  If not, see <https://www.gnu.org/licenses/>.
 */

package daemon

import (
	"context"
	"net/http"
	"sync"

	"github.com/gabstv/httpdigest"
	"github.com/rs/zerolog/log"
	"gitlab.com/moneropay/go-monero/walletrpc"
)

var wallet *walletrpc.Client
var wMutex sync.Mutex
var WalletPrimaryAddress string

func readWalletPrimaryAddress() {
	resp, err := wallet.GetAddress(context.Background(), &walletrpc.GetAddressRequest{AddressIndex: []uint64{0}})
	if err != nil {
		log.Fatal().Err(err).Msg("Startup failure")
	}
	WalletPrimaryAddress = resp.Address
}

func walletConnect() {
	wallet = walletrpc.New(walletrpc.Config{
		Address: Config.rpcAddr,
		Client: &http.Client{
			Transport: httpdigest.New(Config.rpcUser, Config.rpcPass),
		},
	})
	readWalletPrimaryAddress()
}

func Balance(ctx context.Context, indices []uint64) (*walletrpc.GetBalanceResponse, error) {
	wMutex.Lock()
	resp, err := wallet.GetBalance(ctx, &walletrpc.GetBalanceRequest{AddressIndices: indices})
	wMutex.Unlock()
	return resp, err
}

func Transfer(ctx context.Context, r *walletrpc.TransferRequest) (*walletrpc.TransferResponse, error) {
	wMutex.Lock()
	resp, err := wallet.Transfer(ctx, r)
	wMutex.Unlock()
	return resp, err
}

func GetTransfers(ctx context.Context, r *walletrpc.GetTransfersRequest) (*walletrpc.GetTransfersResponse, error) {
	wMutex.Lock()
	resp, err := wallet.GetTransfers(ctx, r)
	wMutex.Unlock()
	return resp, err
}

func GetTransferByTxid(ctx context.Context, r *walletrpc.GetTransferByTxidRequest) (*walletrpc.GetTransferByTxidResponse, error) {
	wMutex.Lock()
	resp, err := wallet.GetTransferByTxid(ctx, r)
	wMutex.Unlock()
	return resp, err
}

func createAddress(ctx context.Context, r *walletrpc.CreateAddressRequest) (*walletrpc.CreateAddressResponse, error) {
	wMutex.Lock()
	resp, err := wallet.CreateAddress(ctx, r)
	wMutex.Unlock()
	return resp, err
}

func SweepAll(ctx context.Context, r *walletrpc.SweepAllRequest) (*walletrpc.SweepAllResponse, error) {
	wMutex.Lock()
	resp, err := wallet.SweepAll(ctx, r)
	wMutex.Unlock()
	return resp, err
}
