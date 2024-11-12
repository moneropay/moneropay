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
	"time"

	"github.com/gabstv/httpdigest"
	"github.com/rs/zerolog/log"
	"gitlab.com/moneropay/go-monero/walletrpc"
)

var wallet *walletrpc.Client
var wMutex sync.Mutex
var WalletPrimaryAddress string

func walletConnect() {
	wallet = walletrpc.New(walletrpc.Config{
		Address: Config.rpcAddr,
		Client: &http.Client{
			Transport: httpdigest.New(Config.rpcUser, Config.rpcPass),
		},
	})
}

func isWalletOpen() bool {
	ctx, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()
	if err := wallet.Store(ctx); err != nil {
		if isWallet, werr := walletrpc.GetWalletError(err); isWallet {
			if werr.Code == walletrpc.ErrNotOpen {
				return false
			}
			err = werr
		}
		log.Fatal().Err(err).Msg("Received erroneous response from monero-wallet-rpc at startup.")
	}
	log.Info().Msg("Wallet was already openned by wallet-rpc.")
	return true
}

func createDefaultWallet() {
	ctx, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()
	if err := wallet.CreateWallet(ctx, &walletrpc.CreateWalletRequest{
		Filename: "wallet", Language: "English",
	}); err != nil {
		if isWallet, werr := walletrpc.GetWalletError(err); isWallet {
			if werr.Code != -21 {
				log.Fatal().Err(werr).Msg("Received erroneous response from monero-wallet-rpc at startup.")
			}
		} else {
			log.Fatal().Err(err).Msg("Received erroneous response from monero-wallet-rpc at startup.")
		}
	}
}

func openDefaultWallet() {
	ctx, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()
	if err := wallet.OpenWallet(ctx, &walletrpc.OpenWalletRequest{Filename: "wallet"}); err != nil {
		if isWallet, werr := walletrpc.GetWalletError(err); isWallet {
			err = werr
		}
		log.Fatal().Err(err).Msg("Received erroneous response from monero-wallet-rpc at startup.")
	}
}

func walletCreateAndOpen() {
	if !isWalletOpen() {
		createDefaultWallet()
		openDefaultWallet()
	}
}

func walletGatherInfo() {
	ctx, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()
	resp, err := wallet.GetAddress(ctx, &walletrpc.GetAddressRequest{AddressIndex: []uint64{0}})
	if err != nil {
		if isWallet, werr := walletrpc.GetWalletError(err); isWallet {
			err = werr
		}
		log.Fatal().Err(err).Msg("Failed to read primary address.")
	}
	WalletPrimaryAddress = resp.Address
}

func Balance(ctx context.Context, indices []uint64) (*walletrpc.GetBalanceResponse, error) {
	wMutex.Lock()
	resp, err := wallet.GetBalance(ctx, &walletrpc.GetBalanceRequest{AddressIndices: indices})
	wMutex.Unlock()
	return resp, err
}

func TransferSplit(ctx context.Context, r *walletrpc.TransferSplitRequest) (*walletrpc.TransferSplitResponse, error) {
	wMutex.Lock()
	resp, err := wallet.TransferSplit(ctx, r)
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

func getHeight(ctx context.Context) (*walletrpc.GetHeightResponse, error) {
	wMutex.Lock()
	resp, err := wallet.GetHeight(ctx)
	wMutex.Unlock()
	return resp, err
}

var cryptonoteDefaultTxSpendableAge uint64 = 10

func getTransferLockStatus(t walletrpc.Transfer) (bool, uint64) {
	locked := true
	eventHeight := t.Height
	// 10 block lock is enforced as a blockchain consensus rule
	if t.Confirmations >= cryptonoteDefaultTxSpendableAge {
		// If the transfer is unlocked compare the block which it unlocked at
		// (t.Height + t.UnlockTime) to the block that caused the last callback
		if t.UnlockTime == 0 || t.UnlockTime-t.Height <= cryptonoteDefaultTxSpendableAge {
			eventHeight += 10
			locked = false
		} else if t.UnlockTime-t.Height <= t.Confirmations {
			eventHeight = t.UnlockTime
			locked = false
		}
	}
	return locked, eventHeight
}
