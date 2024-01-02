# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
