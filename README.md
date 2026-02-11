# MoneroPay

A backend service for receiving, sending and tracking status of Monero payments.

[![MoneroPay.eu](https://moneropay.eu/images/mpay-banner.webp)](https://moneropay.eu/)

MoneroPay provides a simple HTTP API for merchants or individuals who want to accept XMR.
MoneroPay supports optional status updates via HTTP Callbacks.

Documentation of MoneroPay can be found at [https://moneropay.eu](https://moneropay.eu).

Brought to you by:

[![Digilol Software Dev, Hosting & Cybersecurity](https://www.digilol.net/banner-hosting-development.png)](https://www.digilol.net)

## Features

- 0-conf support (enable with `--zero-conf`)
- Timelock handling
- Subaddress based payments
- Partial payments support
- View-only and hot wallet support
- View-only wallet bootstrap mode (enable with `--init-view-only`)
- PostgreSQL and SQLite3 support
- Callbacks when funds arrive and when they unlock

## Quick Start

```bash
# Copy and edit environment variables
cp .env.example .env

# Start with Docker Compose
docker compose up
```

For view-only wallet initialization, uncomment the `INIT_VIEW_ONLY`, `INIT_VIEW_ONLY_NETWORK`, and `DAEMON_ADDRESS` environment variables in `docker-compose.yaml`, then call `GET /keys` once to create the wallet and retrieve the mnemonic seed.

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Service health check |
| GET | `/balance` | Wallet balance |
| POST | `/receive` | Create payment request |
| GET | `/receive/{address}` | Check payment status |
| DELETE | `/receive/{address}` | Remove payment request |
| POST | `/transfer` | Send funds (hot wallet only) |
| GET | `/transfer/{tx_hash}` | Check outgoing TX status |
| GET | `/keys` | One-time key retrieval (view-only init mode) |

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--bind` | `localhost:5000` | Server bind address |
| `--rpc-address` | `http://localhost:18082/json_rpc` | Wallet RPC URL |
| `--postgresql` | - | PostgreSQL connection string |
| `--sqlite` | - | SQLite connection string |
| `--zero-conf` | `false` | Enable 0-conf callbacks |
| `--poll-frequency` | `5s` | Callback polling interval |
| `--init-view-only` | `false` | Bootstrap a new view-only wallet |
| `--init-view-only-network` | `mainnet` | Network: mainnet, testnet, stagenet |
| `--daemon-address` | - | Monero daemon RPC (required for `--init-view-only`) |

## Projects

Example projects that use MoneroPay:

- [MoneroNodo](https://moneronodo.com)
- [Monero ATM Project](https://atm.monero.is)
- [Monero Pagamentos Bot](http://t.me/MoneroPagamentosBot)
- [kernal-donate](https://gitlab.com/kernal/kDonate)
- [XMRpos](https://github.com/Monero-Merchant/monero-merchant)

Do you have a project that uses MoneroPay? Tell us about it and we'll list your project here!

## Contributing

Please prefer [GitLab](https://gitlab.com/moneropay/moneropay/) for opening issues and merge requests.\
Alternatively, you can send patch files via email at [moneropay@kernal.eu](mailto:moneropay@kernal.eu).\
For development related discussions and questions join [#moneropay:kernal.eu](https://matrix.to/#/#moneropay:kernal.eu) Matrix group.
