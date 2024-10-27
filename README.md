# MoneroPay
A backend service for receiving, sending and tracking status of Monero payments.

MoneroPay provides a simple HTTP API for merchants or individuals who want to accept XMR.
MoneroPay supports optional status updates via HTTP Callbacks.

Documentation of MoneroPay can be found at [https://moneropay.eu](https://moneropay.eu).

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

Do you have a project that uses MoneroPay? Tell us about it and we'll list your project here!

## Contributing
Please prefer [GitLab](https://gitlab.com/moneropay/moneropay/) for opening issues and merge requests.\
Alternatively, you can send patch files via email at [moneropay@kernal.eu](mailto:moneropay@kernal.eu).\
For development related discussions and questions join [#moneropay:kernal.eu](https://matrix.to/#/#moneropay:kernal.eu) Matrix group.

## Donate
MoneroPay doesn't receive funds from the Community Crowdfunding System (CCS). Feel free to donate to this address if you're a MoneroPay enjoyer:\
`46VGoe3bKWTNuJdwNjjr6oGHLVtV1c9QpXFP9M2P22bbZNU7aGmtuLe6PEDRAeoc3L7pSjfRHMmqpSF5M59eWemEQ2kwYuw`

If you would like to leave a message, you can use [kDonate](https://donate.kernal.eu).
