/*
 * MoneroPay is a Monero payment processor.
 * Copyright (C) 2026 Laurynas Četyrkinas <laurynas@digilol.net>
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
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.com/moneropay/go-monero/walletrpc"
	"gitlab.com/moneropay/moneropay/v2/pkg/model"
	"gitlab.com/moneropay/moneropay/v2/pkg/xmrkey"
)

// getDaemonHeight queries the Monero daemon for the current blockchain height.
func (d *Daemon) getDaemonHeight(ctx context.Context) (uint64, error) {
	reqBody := []byte(`{"jsonrpc":"2.0","id":"0","method":"get_info"}`)
	req, err := http.NewRequestWithContext(ctx, "POST", d.config.DaemonAddr, bytes.NewReader(reqBody))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Result struct {
			Height uint64 `json:"height"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	if result.Error != nil {
		return 0, fmt.Errorf("daemon error: %s", result.Error.Message)
	}
	return result.Result.Height, nil
}

// getNetworkByte returns the network byte for the configured network.
// Config validation ensures the network value is valid before this is called.
func (d *Daemon) getNetworkByte() byte {
	switch d.config.InitViewOnlyNetwork {
	case "testnet":
		return xmrkey.NetworkTestnet
	case "stagenet":
		return xmrkey.NetworkStagenet
	default:
		return xmrkey.NetworkMainnet
	}
}

// initViewOnlyWallet generates new keys but does NOT create the wallet yet.
// The wallet is created when GetOneTimeKeys is called.
func (d *Daemon) initViewOnlyWallet() {
	key, err := xmrkey.NewKey(d.getNetworkByte())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to generate Monero keys")
	}

	mnemonic, err := key.Mnemonic("english")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to generate mnemonic")
	}

	d.keysMutex.Lock()
	d.pendingKey = key
	d.pendingMnemonic = mnemonic
	d.walletCreated = false
	d.keysMutex.Unlock()
}

// IsKeysPending returns true if we're in view-only init mode and keys haven't been retrieved yet.
func (d *Daemon) IsKeysPending() bool {
	d.keysMutex.Lock()
	defer d.keysMutex.Unlock()
	return d.pendingKey != nil && !d.walletCreated
}

// IsKeysConsumed returns true if keys were already retrieved.
func (d *Daemon) IsKeysConsumed() bool {
	d.keysMutex.Lock()
	defer d.keysMutex.Unlock()
	return d.walletCreated
}

// IsViewOnlyInit returns true if server was started with --init-view-only.
func (d *Daemon) IsViewOnlyInit() bool {
	d.keysMutex.Lock()
	defer d.keysMutex.Unlock()
	return d.pendingKey != nil || d.walletCreated
}

// GetOneTimeKeys creates the wallet and returns all key information exactly once, then clears it.
// Caller should check IsViewOnlyInit() and IsKeysConsumed() before calling.
func (d *Daemon) GetOneTimeKeys() (*model.KeysResponse, error) {
	d.keysMutex.Lock()
	defer d.keysMutex.Unlock()

	// Create the view-only wallet using generate_from_keys RPC
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get current height from daemon to use as restore height.
	// Since this is a brand new wallet, there's no transaction history to scan.
	height, err := d.getDaemonHeight(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get daemon height: %w", err)
	}
	log.Info().Uint64("height", height).Msg("Creating view-only wallet at current height")

	_, err = d.wallet.GenerateFromKeys(ctx, &walletrpc.GenerateFromKeysRequest{
		Filename:        "wallet",
		Address:         d.pendingKey.Address,
		ViewKey:         hex.EncodeToString(d.pendingKey.PrivateViewKey[:]),
		Password:        "",
		AutosaveCurrent: false,
		RestoreHeight:   height,
	})
	if err != nil {
		return nil, err
	}

	// Build response
	keys := &model.KeysResponse{
		Mnemonic:        d.pendingMnemonic,
		RestoreHeight:   height,
		SpendKeyPublic:  hex.EncodeToString(d.pendingKey.PublicSpendKey[:]),
		SpendKeyPrivate: hex.EncodeToString(d.pendingKey.PrivateSpendKey[:]),
		ViewKeyPublic:   hex.EncodeToString(d.pendingKey.PublicViewKey[:]),
		ViewKeyPrivate:  hex.EncodeToString(d.pendingKey.PrivateViewKey[:]),
	}

	// Clear keys from memory
	d.pendingKey = nil
	d.pendingMnemonic = ""
	d.walletCreated = true

	// Complete daemon initialization before returning
	d.CompleteViewOnlyInit()

	return keys, nil
}
