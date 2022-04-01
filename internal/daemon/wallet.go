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
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gabstv/httpdigest"
	"gitlab.com/moneropay/go-monero/walletrpc"
)

var wallet *walletrpc.Client
var wMutex sync.Mutex
var WalletPrimaryAddress string

func readWalletPrimaryAddress() {
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	resp, err := wallet.GetAddress(ctx, &walletrpc.GetAddressRequest{AddressIndex: []uint64{0}})
	defer cancel()
	if err != nil {
		log.Fatal(err)
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

func Balance(indices []uint64) (*walletrpc.GetBalanceResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	wMutex.Lock()
	resp, err := wallet.GetBalance(ctx, &walletrpc.GetBalanceRequest{AddressIndices: indices})
	defer cancel()
	wMutex.Unlock()
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func Transfer(r *walletrpc.TransferRequest) (*walletrpc.TransferResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	wMutex.Lock()
	resp, err := wallet.Transfer(ctx, r)
	defer cancel()
	wMutex.Unlock()
	if err != nil {
		return nil, err
	}
	return resp, err
}

func GetTransfers(r *walletrpc.GetTransfersRequest) (*walletrpc.GetTransfersResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	wMutex.Lock()
	resp, err := wallet.GetTransfers(ctx, r)
	defer cancel()
	wMutex.Unlock()
	if err != nil {
		return nil, err
	}
	return resp, err
}

func GetTransferByTxid(r *walletrpc.GetTransferByTxidRequest) (*walletrpc.GetTransferByTxidResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	wMutex.Lock()
	resp, err := wallet.GetTransferByTxid(ctx, r)
	defer cancel()
	wMutex.Unlock()
	if err != nil {
		return nil, err
	}
	return resp, err
}

func createAddress(r *walletrpc.CreateAddressRequest) (*walletrpc.CreateAddressResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	wMutex.Lock()
	resp, err := wallet.CreateAddress(ctx, r)
	defer cancel()
	wMutex.Unlock()
	if err != nil {
		return nil, err
	}
	return resp, nil
}
