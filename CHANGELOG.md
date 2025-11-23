# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.8.1] - 2025-11-24
### Fixed
- Updated dependencies to fix a bug related to go-monero library's TransferSplitResponse reported by EgeBalci.

## [2.8.0] - 2025-09-11
### Added
- When 0-conf is enabled the `GET /receive` endpoint now returns mempool transactions.

## [2.7.1] - 2024-12-28
### Fixed
- The `/health` endpoint now handles timeouts properly instead of hanging. This bug was introduced in 2.7.0.

## [2.7.0] - 2024-11-24
### Added
- New `DELETE /receive/{address}` endpoint.
- New wallet auto-creation feature. If no wallet file is provided on `monero-wallet-rpc` side. We will attempt to create a wallet via RPC calls.

### Fixed
- `moneropay-port-db` tool was misisng `mempool_seen` table migration.
- `UnlockTime` argument was removed from `TransferSplit` function, because non 0 values got deprecated on Monero side. And our default was set to 10.

### Changed
- Docker Compose files to adapt the new wallet auto-creation feature.
- Docker Compose files to enable 0-conf.
- Health check now uses `refresh` RPC call instead of `get_height`.

### Removed
- `transfer-mixin` CLI argument that was used in `TransferSplit` function as `RingSize: daemon.Config.TransferMixin + 1` argument.
- `transfer-unlock-time` CLI argument as it was deprecated on Monero side.

## [2.6.0] - 2024-10-27
### Added
- Added optional 0-conf support. When enabled it causes 3 callbacks per tx to be sent: 0-conf, 1-conf and unlock.
- Added configuration option for setting the interval of incoming payment checks. By default it is 5 seconds (`5s`).

### Fixed
- Shortened the sleep and changed the workflow of the callback thread. Callbacks should be more timely in theory.

## [2.5.1] - 2024-09-30
### Changed
- Added a retry mechanism for the initial connection to monero-wallet-rpc. Improved the logging that happens in this stage.
- Updated dependencies and monero-wallet-rpc version in docker-compose.yaml. Also removed the logging arguments to monero-wallet-rpc in docker-compose.yaml

## [2.5.0] - 2024-01-02
### Added
- Added `tx_hash_list` field in `POST /transfer` response. If there are multiple TX hashes, then `tx_hash` field will contain the first hash in the list. `tx_hash` field will be removed the next major release.

### Changed
- `POST /transfer` now uses the wallet RPC call `transfer_split` instead of `transfer`.

### Fixed
- Typo `postgresq` in docker-compose.yaml.

## [2.4.0] - 2023-10-20
### Added
- SQLite3 support (special thanks to [recanman](http://recanman7nly4wwc5f2t2h55jnxsr7wo664o3lsydngwetvrguz4esid.onion/) for testing and implementing migration files).

### Changed
- Dockerfile build procedure. [techknowlogick/xgo](https://github.com/techknowlogick/xgo) is now used to build the Docker images due to CGO dependency ([mattn/go-sqlite3](http://mattn.github.io/go-sqlite3/)).
- Bumped monero-wallet-rpc version to v0.18.3.1 in docker-compose.yaml.
- Switched from [jackc/pgx](https://github.com/jackc/pgx) driver to standard library's [database/sql](https://pkg.go.dev/database/sql).

### Fixed
- Rows are now properly closed after scanning.

## [2.3.0] - 2023-07-11
### Added
- Export callback response struct under `/pkg/model`.

## [2.2.2] - 2023-04-02
### Added
- Exported API structs under `/pkg/model`.

### Fixed
- Fixed bug where MoneroPay attempts to callback when callback URL isn't supplied.

## [2.2.1] - 2023-02-18
### Fixed
- Upgraded [pgx](https://github.com/jackc/pgx) PostgreSQL driver library to v5 from v4. Dead connection and broken pipe handling was added.

## [2.2.0] - 2022-09-01
### Added
- Added logging using zerolog. Two formats available: pretty and JSON.

### Changed
- Callbacks are now also sent when the transfers unlock. The callback JSON data now contains locked parameter of boolean type.
- Transactions for GET /receive response now contain locked field. Indicates the transfer's lock status
- Updated Dockerfile and docker-compose.yaml to support Monero v0.18.1.0.

### Fixed
- Fixed the transfer bug. MoneroPay no longer relies on the wallet-rpc for unlocked balances of subaddresses because they can change on outgoing transfer. The accounting of transfers is now done in the database.
- Debloated migration files.

## [2.1.0] - 2022-05-16
### Added
- Implemented database migrations. v1 <-> v2 and v2 <-> v2.1.0

### Fixed
- Improved the way we handle timeouts. We now use a middleware and care about router's context.
- Fixed some more bugs.
