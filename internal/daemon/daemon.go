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

// Package daemon implements the MoneroPay backend services.
// It consists of two main components:
// - Wallet RPC client for Monero wallet operations
// - Callback runner for payment notification polling
package daemon

import (
	"context"
	"database/sql"
	"sync"

	"github.com/rs/zerolog/log"
	"gitlab.com/moneropay/go-monero/walletrpc"
	"gitlab.com/moneropay/moneropay/v2/pkg/xmrkey"
)

const Version = "2.9.0"

// Daemon encapsulates all MoneroPay backend services and state.
type Daemon struct {
	config *Config
	wallet *walletrpc.Client
	db     *sql.DB

	wMutex sync.Mutex // protects wallet RPC calls

	// Callback runner state
	lastCallbackHeight uint64
	lastSeenHeight     uint64
	shutdownCancel     context.CancelFunc

	// View-only mode state
	keysMutex       sync.Mutex
	pendingKey      *xmrkey.Key
	pendingMnemonic string
	walletCreated   bool

	// WalletPrimaryAddress is the wallet's primary address, set after initialization.
	WalletPrimaryAddress string

	// isViewOnly is true if the wallet has no spend key.
	isViewOnly bool
}

// New creates a new Daemon from the given configuration.
func New(cfg *Config) *Daemon {
	return &Daemon{
		config: cfg,
	}
}

// Config returns the daemon's configuration.
func (d *Daemon) Config() *Config {
	return d.config
}

// Run initializes and starts all daemon services.
// It handles two modes:
// 1. Normal mode: Creates/opens wallet, connects DB, starts callback runner
// 2. View-only init mode: Opens existing view-only wallet or generates keys for new one
func (d *Daemon) Run() {
	d.connectWallet()

	if d.config.InitViewOnly {
		d.runViewOnlyMode()
		return
	}

	d.walletCreateAndOpen()
	d.completeInit()
}

// runViewOnlyMode handles the view-only wallet initialization flow.
// If a wallet exists, it opens and verifies it's view-only.
// If no wallet exists, it generates keys and waits for /keys endpoint.
func (d *Daemon) runViewOnlyMode() {
	exists, err := d.tryOpenExistingWallet()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to check for existing wallet")
	}

	if exists {
		// Wallet exists - verify it's view-only
		isViewOnly, err := d.isViewOnlyWallet()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to check wallet type")
		}
		if !isViewOnly {
			log.Fatal().Msg("Existing wallet is not view-only. Remove --init-view-only flag or use a view-only wallet.")
		}
		log.Info().Msg("Opened existing view-only wallet")
		d.completeInit()
		return
	}

	// No wallet exists - generate keys and wait for /keys endpoint
	log.Info().Msg("No wallet found, generating keys for new view-only wallet")
	d.initViewOnlyWallet()
}

// completeInit finishes initialization after wallet is ready.
func (d *Daemon) completeInit() {
	d.gatherWalletInfo()
	d.checkViewOnly()
	d.connectDB()
	d.readLastCallbackHeight(context.Background())
	d.runMigrations()

	ctx, cancel := context.WithCancel(context.Background())
	d.shutdownCancel = cancel
	go d.callbackRunner(ctx)
}

// CompleteViewOnlyInit is called after keys are retrieved to finish initialization.
func (d *Daemon) CompleteViewOnlyInit() {
	d.completeInit()
}

// checkViewOnly checks if the wallet is view-only and caches the result.
func (d *Daemon) checkViewOnly() {
	viewOnly, err := d.isViewOnlyWallet()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to check if wallet is view-only")
		return
	}
	d.isViewOnly = viewOnly
	if viewOnly {
		log.Info().Msg("Wallet is view-only, transfers are disabled")
	}
}

// IsViewOnly returns true if the wallet has no spend key.
func (d *Daemon) IsViewOnly() bool {
	return d.isViewOnly
}

// Shutdown gracefully stops all daemon services.
func (d *Daemon) Shutdown() {
	if d.shutdownCancel != nil {
		d.shutdownCancel()
	}
}
