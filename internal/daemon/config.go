/*
 * MoneroPay is a Monero payment processor.
 * Copyright (C) 2022 İrem Kuyucu <siren@kernal.eu>
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
	"time"

	"github.com/jnovack/flag"
	"github.com/rs/zerolog/log"
)

// Config holds all daemon configuration.
type Config struct {
	// Server
	BindAddr string

	// Wallet RPC connection
	RPCAddr string
	RPCUser string
	RPCPass string

	// Transfer settings
	TransferMixin      uint64 // Deprecated
	TransferPriority   uint
	TransferUnlockTime uint64 // Deprecated

	// Database (mutually exclusive)
	PostgresCS string
	SQLiteCS   string

	// Logging
	LogFormat string

	// Callback settings
	ZeroConf bool
	PollFreq time.Duration

	// View-only initialization
	InitViewOnly        bool
	InitViewOnlyNetwork string
	DaemonAddr          string
}

// LoadConfig loads configuration from flags and environment variables.
func LoadConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.BindAddr, "bind", "localhost:5000", "Bind address:port for moneropayd")
	flag.StringVar(&cfg.RPCAddr, "rpc-address", "http://localhost:18082/json_rpc", "Wallet RPC server address")
	flag.StringVar(&cfg.RPCUser, "rpc-username", "", "Username for monero-wallet-rpc")
	flag.StringVar(&cfg.RPCPass, "rpc-password", "", "Password for monero-wallet-rpc")
	flag.Uint64Var(&cfg.TransferMixin, "transfer-mixin", 0, "Deprecated and ignored, will be removed the next major release (3.0.0)")
	flag.UintVar(&cfg.TransferPriority, "transfer-priority", 0, "Set a priority for transactions")
	flag.Uint64Var(&cfg.TransferUnlockTime, "transfer-unlock-time", 0, "Deprecated and ignored, will be removed the next major release (3.0.0)")
	flag.StringVar(&cfg.PostgresCS, "postgresql", "postgresql://moneropay:s3cret@localhost:5432/moneropay", "PostgreSQL connection string")
	flag.StringVar(&cfg.SQLiteCS, "sqlite", "", "SQLite3 connection string")
	flag.StringVar(&cfg.LogFormat, "log-format", "pretty", "Log format (pretty or json)")
	flag.BoolVar(&cfg.ZeroConf, "zero-conf", false, "Enable 0-conf mode. Sends 3 callbacks (0-conf, 1-conf, 10-conf)")
	flag.DurationVar(&cfg.PollFreq, "poll-frequency", 5*time.Second, "Interval for checking new incoming and pool payments")
	flag.BoolVar(&cfg.InitViewOnly, "init-view-only", false, "Initialize a new view-only wallet")
	flag.StringVar(&cfg.InitViewOnlyNetwork, "init-view-only-network", "mainnet", "Network for view-only wallet: mainnet, testnet, or stagenet")
	flag.StringVar(&cfg.DaemonAddr, "daemon-address", "", "Monero daemon address")
	flag.String(flag.DefaultConfigFlagname, "", "Path to configuration file")
	flag.Parse()

	cfg.validate()
	return cfg
}

// validate checks that configuration values are valid.
func (c *Config) validate() {
	switch c.InitViewOnlyNetwork {
	case "mainnet", "testnet", "stagenet":
	default:
		log.Fatal().Str("network", c.InitViewOnlyNetwork).
			Msg("Invalid network. Must be: mainnet, testnet, or stagenet")
	}

	if c.InitViewOnly && c.DaemonAddr == "" {
		log.Fatal().Msg("--daemon-address is required when using --init-view-only")
	}
}
