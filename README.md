# MoneroPay
A backend service for receiving, sending and tracking status of Monero payments.

MoneroPay provides a simple HTTP API for merchants or individuals who want to accept XMR.
MoneroPay supports optional status updates via HTTP Callbacks.

Documentation of MoneroPay can be found at [https://moneropay.eu](https://moneropay.eu).

Brought to you by:

[![Digilol Software Dev, Hosting & Cybersecurity](https://www.digilol.net/banner-hosting-development.png)](https://www.digilol.net)



## Features
- 0-conf (enable with `--zero-conf=true`)
- Timelock handling.
- Subaddress based.
- Partial payments support.
- View-only and hot wallet support.
- PostgreSQL and SQLite3 support.
- Callbacks when funds arrive and when they unlock.

## Projects
Example projects that use MoneroPay:

- [MoneroNodo](https://moneronodo.com)
- [Monero ATM Project](https://atm.monero.is)
- [Monero Pagamentos Bot](http://t.me/MoneroPagamentosBot)
- [kernal-donate](https://gitlab.com/kernal/kDonate)

Do you have a project that uses MoneroPay? Tell us about it and we'll list your project here!

## Contributing
Please prefer [GitLab](https://gitlab.com/moneropay/moneropay/) for opening issues and merge requests.\
Alternatively, you can send patch files via email at [moneropay@kernal.eu](mailto:moneropay@kernal.eu).\
For development related discussions and questions join [#moneropay:kernal.eu](https://matrix.to/#/#moneropay:kernal.eu) Matrix group.