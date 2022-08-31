# MoneroPay
A backend service for receiving, sending and tracking status of Monero payments.

MoneroPay provides a simple HTTP API for merchants or individuals who want to accept XMR.
MoneroPay supports optional status updates via HTTP Callbacks.

[See here](https://donate.kernal.eu) for a donation page utilizing MoneroPay.

Documentation on MoneroPay can be found [here](https://moneropay.eu).

## Endpoints
| Method | URI                            | Input                                                                                 |
| :----: | ------------------------------ | ------------------------------------------------------------------------------------- |
| `GET`  | /balance                       |                                                                                       |
| `GET`  | /health                        |                                                                                       |
| `POST` | /receive                       | `{"amount": 123000000, "description": "Stickers", "callback_url": "http://merchant"}` |
| `GET`  | /receive/:subaddress?min=&max= |                                                                                       |
| `POST` | /transfer                      | `{"destinations": [{"amount": 1337000000, "address": "47stn..."}]}`                   |
| `GET`  | /transfer/:tx_hash             |                                                                                       |

### Balance
#### Request
```sh
curl -s -X GET "${endpoint}/balance"
```
#### Response
##### 200 (Success)
```json
{
  "total": 2513444800,
  "unlocked": 800000000,
}
```

### Health
#### Request
```sh
curl -s -X GET "${endpoint}/health"
```
#### Response
##### 200 (Success)
```json
{
  "status": 200,
  "services": {
    "walletrpc": true,
    "postgresql": true
  }
}
```

### Receive
#### Request
```sh
curl -s -X POST "${endpoint}/receive" \
  -H 'Content-Type: application/json' \
  -d '{"amount": 123000000, "description": "Server expenses", "callback_url": "http://merchant"}'
```
#### Response
##### 200 (Success)
```json
{
  "address": "84WsptnLmjTYQjm52SMkhQWsepprkcchNguxdyLkURTSW1WLo3tShTnCRvepijbc2X8GAKPGxJK9hfQhLHzoKSxh7y8Yqrg",
  "amount": 123000000,
  "description": "Server expenses",
  "created_at": "2022-07-18T11:54:49.780542861Z"
}
```

### Receive tracking
#### Request
```sh
curl -s -X GET "${endpoint}/receive/${address}?min=${min_height}&max=${max_height}"
```
#### Response
##### 200 (Success)
```json
{
  "amount": {
    "expected": 1,
    "covered": {
      "total": 200000000,
      "unlocked": 200000000
    }
  },
  "complete": true,
  "description": "Donation to Kernal",
  "created_at": "2022-07-11T19:04:24.574583Z",
  "transactions": [
    {
      "amount": 200000000,
      "confirmations": 10,
      "double_spend_seen": false,
      "fee": 9200000,
      "height": 2402648,
      "timestamp": "2022-07-11T19:19:05Z",
      "tx_hash": "0c9a7b40b15596fa9a06ba32463a19d781c075120bb59ab5e4ed2a97ab3b7f33",
      "unlock_time": 0,
      "locked": false
    }
  ]
}
```

### Transfer
#### Request
```sh
curl -s -X POST "${endpoint}/transfer" \
	-H 'Content-Type: application/json' \
	-d '{"destinations": [{"amount": 1337000000, "address": "47stn..."}]}'
```
#### Response
##### 200 (Success)
```json
{
  "amount": 1337000000,
  "fee": 87438594,
  "tx_hash": "5ca34...",
  "destinations": [
    {
      "amount": 1337,
      "address": "47stn..."
    }
  ]
}
```

### Transfer tracking
#### Request
```sh
curl -s -X GET "${endpoint}/transfer/${tx_hash}"
```
#### Response
##### 200 (Success)
```json
{
  "amount": 79990000,
  "fee": 9110000,
  "state": "completed",
  "transfer": [
    {
      "amount": 79990000,
      "address": "453biCQpM6oSSr7jgTwmtC9YfiXUWZY1wEfSZJD4r6rf7mPqPj8NZpp7WYpAHVq7p69SYa1B1zMN6SeRc8exYi1WEenqu2c"
    }
  ],
  "confirmations": 15,
  "double_spend_seen": false,
  "height": 2407445,
  "timestamp": "2022-07-18T11:37:50Z",
  "unlock_time": 10,
  "tx_hash": "cf448effb86f24f81476c0012a6636700488e13accd91f8f43302ae90fed25ce"
}
```

### Callback payload
```json
{
  "amount": {
    "expected": 0,
    "covered": {
      "total": 200000000,
      "unlocked": 200000000
    }
  },
  "complete": true,
  "description": "Donation to Kernal",
  "created_at": "2022-07-11T19:04:24.574583Z",
  "transaction": {
    "amount": 200000000,
    "confirmations": 10,
    "double_spend_seen": false,
    "fee": 9200000,
    "height": 2402648,
    "timestamp": "2022-07-11T19:19:05Z",
    "tx_hash": "0c9a7b40b15596fa9a06ba32463a19d781c075120bb59ab5e4ed2a97ab3b7f33",
    "unlock_time": 0,
    "locked": false
  }
}
```

## Usage
```
$ ./moneropay -h
Usage of ./moneropay:
  -bind="localhost:5000": Bind address:port for moneropayd
  -config="": Path to configuration file
  -log-format="pretty": Log format (pretty or json)
  -postgresql="postgresql://moneropay:s3cret@localhost:5432/moneropay": PostgreSQL connection string
  -rpc-address="http://localhost:18082/json_rpc": Wallet RPC server address
  -rpc-password="": Password for monero-wallet-rpc
  -rpc-username="": Username for monero-wallet-rpc
  -transfer-mixin=8: Number of outputs from the blockchain to mix with (0 means no mixing)
  -transfer-priority=0: Set a priority for transactions
  -transfer-unlock-time=10: Number of blocks before the monero can be spent (0 to not add a lock)
```
Environment variables are also supported.
```sh
#!/bin/sh
export RPC_ADDRESS='http://localhost:18083/json_rpc'
export RPC_USERNAME='kernal'
export RPC_PASSWORD='s3cure'
export POSTGRESQL='postgresql://moneropay:s3cret@localhost:5432/moneropay'
./moneropay
```

## Contributing
Issues and merge requests are only checked on [GitLab](https://gitlab.com/moneropay/moneropay/).\
Alternatively, you can send patch files via email at [moneropay@kernal.eu](mailto:moneropay@kernal.eu).\
For development related discussions and questions join [#moneropay:kernal.eu](https://matrix.to/#/#moneropay:kernal.eu) Matrix group.
