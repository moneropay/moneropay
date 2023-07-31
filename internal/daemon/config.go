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

import "github.com/namsral/flag"

type config struct {
	BindAddr string
	rpcAddr string
	rpcUser string
	rpcPass string
	TransferMixin uint64
	TransferPriority uint
	TransferUnlockTime uint64
	dbCS string
	dbDriver string
	sqliteFilePath string
	logFormat string
}

var Config config

func loadConfig() {
	flag.StringVar(&Config.BindAddr, "bind", "localhost:5000", "Bind address:port for moneropayd")
	flag.StringVar(&Config.rpcAddr, "rpc-address", "http://localhost:18082/json_rpc", "Wallet RPC server address")
	flag.StringVar(&Config.rpcUser, "rpc-username", "", "Username for monero-wallet-rpc")
	flag.StringVar(&Config.rpcPass, "rpc-password", "", "Password for monero-wallet-rpc")
	flag.Uint64Var(&Config.TransferMixin, "transfer-mixin", 8, "Number of outputs from the blockchain to mix with (0 means no mixing)")
	flag.UintVar(&Config.TransferPriority, "transfer-priority", 0, "Set a priority for transactions")
	flag.Uint64Var(&Config.TransferUnlockTime, "transfer-unlock-time", 10, "Number of blocks before the monero can be spent (0 to not add a lock)")
	flag.StringVar(&Config.dbDriver, "db-driver", "postgres", "The database driver to use. Can be postgres or sqlite3.") // TODO: validate
	flag.StringVar(&Config.dbCS, "db-conn-string", "postgresql://moneropay:s3cret@localhost:5432/moneropay", "Database connection string")
	flag.StringVar(&Config.sqliteFilePath, "sqlite3-file-path", "./db.sqlite3", "For sqlite3, the path to the database file (used separately from db-conn-string).")
	flag.StringVar(&Config.logFormat, "log-format", "pretty", "Log format (pretty or json)")
	flag.String(flag.DefaultConfigFlagname, "", "Path to configuration file")
	flag.Parse()
}
