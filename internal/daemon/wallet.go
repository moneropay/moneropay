/*
 * MoneroPay is a Monero payment processor.
 * Copyright (C) 2026 Laurynas Četyrkinas <laurynas@digilol.net>
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
	"time"

	"github.com/gabstv/httpdigest"
	"github.com/rs/zerolog/log"
	"gitlab.com/moneropay/go-monero/walletrpc"
)

// connectWallet establishes connection to the Monero wallet RPC.
func (d *Daemon) connectWallet() {
	d.wallet = walletrpc.New(walletrpc.Config{
		Address: d.config.RPCAddr,
		Client: &http.Client{
			Transport: httpdigest.New(d.config.RPCUser, d.config.RPCPass),
		},
	})
}

// isWalletOpen checks if a wallet is already open in the RPC.
func (d *Daemon) isWalletOpen() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := d.wallet.Store(ctx); err != nil {
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

// createDefaultWallet creates a new wallet named "wallet".
func (d *Daemon) createDefaultWallet() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := d.wallet.CreateWallet(ctx, &walletrpc.CreateWalletRequest{
		Filename: "wallet", Language: "English",
	}); err != nil {
		if isWallet, werr := walletrpc.GetWalletError(err); isWallet {
			if werr.Code != -21 { // Wallet file already exists
				log.Fatal().Err(werr).Msg("Received erroneous response from monero-wallet-rpc at startup.")
			}
		} else {
			log.Fatal().Err(err).Msg("Received erroneous response from monero-wallet-rpc at startup.")
		}
	}
}

// openDefaultWallet opens the wallet named "wallet".
func (d *Daemon) openDefaultWallet() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := d.wallet.OpenWallet(ctx, &walletrpc.OpenWalletRequest{Filename: "wallet"}); err != nil {
		if isWallet, werr := walletrpc.GetWalletError(err); isWallet {
			err = werr
		}
		log.Fatal().Err(err).Msg("Received erroneous response from monero-wallet-rpc at startup.")
	}
}

// walletCreateAndOpen ensures a wallet is open, creating one if necessary.
func (d *Daemon) walletCreateAndOpen() {
	if !d.isWalletOpen() {
		d.createDefaultWallet()
		d.openDefaultWallet()
	}
}

// gatherWalletInfo retrieves and stores the wallet's primary address.
// Retries up to 60 times (60 seconds) as wallet may still be initializing.
func (d *Daemon) gatherWalletInfo() {
	var resp *walletrpc.GetAddressResponse
	var err error

	for i := 0; i < 60; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		resp, err = d.wallet.GetAddress(ctx, &walletrpc.GetAddressRequest{AddressIndex: []uint64{0}})
		cancel()

		if err == nil {
			d.WalletPrimaryAddress = resp.Address
			return
		}

		log.Debug().Err(err).Int("attempt", i+1).Msg("Waiting for wallet RPC to become available")
		time.Sleep(1 * time.Second)
	}

	if isWallet, werr := walletrpc.GetWalletError(err); isWallet {
		err = werr
	}
	log.Fatal().Err(err).Msg("Failed to read primary address after retries.")
}

// Balance returns the wallet balance for the given address indices.
func (d *Daemon) Balance(ctx context.Context, indices []uint64) (*walletrpc.GetBalanceResponse, error) {
	d.wMutex.Lock()
	defer d.wMutex.Unlock()
	return d.wallet.GetBalance(ctx, &walletrpc.GetBalanceRequest{AddressIndices: indices})
}

// TransferSplit sends funds to one or more destinations.
func (d *Daemon) TransferSplit(ctx context.Context, r *walletrpc.TransferSplitRequest) (*walletrpc.TransferSplitResponse, error) {
	d.wMutex.Lock()
	defer d.wMutex.Unlock()
	return d.wallet.TransferSplit(ctx, r)
}

// GetTransfers retrieves transfers matching the given criteria.
func (d *Daemon) GetTransfers(ctx context.Context, r *walletrpc.GetTransfersRequest) (*walletrpc.GetTransfersResponse, error) {
	d.wMutex.Lock()
	defer d.wMutex.Unlock()
	return d.wallet.GetTransfers(ctx, r)
}

// GetTransferByTxid retrieves a specific transfer by transaction ID.
func (d *Daemon) GetTransferByTxid(ctx context.Context, r *walletrpc.GetTransferByTxidRequest) (*walletrpc.GetTransferByTxidResponse, error) {
	d.wMutex.Lock()
	defer d.wMutex.Unlock()
	return d.wallet.GetTransferByTxid(ctx, r)
}

// createAddress creates a new subaddress for receiving payments.
func (d *Daemon) createAddress(ctx context.Context, r *walletrpc.CreateAddressRequest) (*walletrpc.CreateAddressResponse, error) {
	d.wMutex.Lock()
	defer d.wMutex.Unlock()
	return d.wallet.CreateAddress(ctx, r)
}

// getHeight returns the current blockchain height as seen by the wallet.
func (d *Daemon) getHeight(ctx context.Context) (*walletrpc.GetHeightResponse, error) {
	d.wMutex.Lock()
	defer d.wMutex.Unlock()
	return d.wallet.GetHeight(ctx)
}

// refresh triggers a wallet refresh and returns the result.
func (d *Daemon) refresh(ctx context.Context) (*walletrpc.RefreshResponse, error) {
	d.wMutex.Lock()
	defer d.wMutex.Unlock()
	return d.wallet.Refresh(ctx, &walletrpc.RefreshRequest{})
}

// isViewOnlyWallet checks if the currently open wallet is view-only.
// Returns true if the wallet has no spend key (view-only), false if it has a spend key (full wallet).
func (d *Daemon) isViewOnlyWallet() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := d.wallet.QueryKey(ctx, &walletrpc.QueryKeyRequest{
		KeyType: string(walletrpc.QueryKeySpend),
	})
	if err != nil {
		if isWallet, werr := walletrpc.GetWalletError(err); isWallet {
			// Error code -29 means "watch-only wallet" (no spend key available)
			if werr.Code == -29 {
				return true, nil
			}
			return false, werr
		}
		return false, err
	}
	// If we successfully got the spend key, it's a full wallet
	return false, nil
}

// tryOpenExistingWallet attempts to open the default wallet.
// Returns true if wallet was opened, false if it doesn't exist.
func (d *Daemon) tryOpenExistingWallet() (bool, error) {
	if d.isWalletOpen() {
		return true, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := d.wallet.OpenWallet(ctx, &walletrpc.OpenWalletRequest{Filename: "wallet"})
	if err != nil {
		if isWallet, werr := walletrpc.GetWalletError(err); isWallet {
			// Error code -1 means wallet file doesn't exist
			if werr.Code == -1 {
				return false, nil
			}
			return false, werr
		}
		return false, err
	}
	return true, nil
}

// checkFatalWalletError checks if the error indicates the wallet is no longer open.
// If so, it exits the process to trigger a container restart.
func checkFatalWalletError(err error) {
	if isWallet, werr := walletrpc.GetWalletError(err); isWallet {
		if werr.Code == walletrpc.ErrNotOpen {
			log.Fatal().Err(werr).Msg("Wallet RPC has no wallet open.")
		}
	}
}

// cryptonoteDefaultTxSpendableAge is the number of confirmations required for funds to be spendable.
const cryptonoteDefaultTxSpendableAge uint64 = 10

// getTransferLockStatus determines if a transfer is locked and returns its event height.
func getTransferLockStatus(t walletrpc.Transfer) (locked bool, eventHeight uint64) {
	locked = true
	eventHeight = t.Height

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
