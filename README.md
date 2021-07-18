# MoneroPay API
API for receiving, sending and tracking Monero payments.

MoneroPay provides a simple backend service for merchants or individuals accepting XMR.
Optionally, it will check for new incoming transfers and callback provided endpoint.

## Endpoints
| Method | URI                  | Input                                                                      |
| :----: | -------------------- | -------------------------------------------------------------------------- |
| `GET`  | /balance             |                                                                            |
| `GET`  | /health              |                                                                            |
| `POST` | /receive             | `'amount=123000000' 'description=Stickers' 'callback_url=http://merchant'` |
| `GET`  | /receive/:subaddress |                                                                            |
| `POST` | /transfer            | `{"destinations": [{"amount": 1337000000, "address": "47stn..."}]}`        |
| `GET`  | /transfer/:tx_hash   |                                                                            |

## Balance
### Request
```sh
curl -s -X GET "${endpoint}/balance"
```
### Response
### 200 (Success)
```jsonc
{
  "total": 2513444800,
  "unlocked": 800000000,
  "locked": 1713444800
}
```

## Health
### Request
```sh
curl -s -X GET "${endpoint}/health"
```
### Response
#### 200 (Success)
```jsonc
{
  "status": 200,
  "services": {
    "walletrpc": true,
    "postgresql": true
  }
}
```

## Receive
### Request
```sh
curl -s -X POST "${endpoint}/receive"
  -d 'amount=123000000' # uint64 (required) - Amount to expect in XMR atomic units.
  -d 'description=Server expenses' # string - The description for the order.
  -d 'callback_url=http://merchant' # string - Callback on incoming transfers.
```
### Response
### 200 (Success)
```jsonc
{
  "address": "84WsptnLmjTYQjm52SMkhQWsepprkcchNguxdyLkURTSW1WLo3tShTnCRvepijbc2X8GAKPGxJK9hfQhLHzoKSxh7y8Yqrg",
  "amount": 123000000,
  "description": "Server expenses",
  "created_at": "2021-07-18T11:54:49.780542861Z"
}
```

## Receive tracking
### Request
```sh
curl -s -X GET "${endpoint}/receive/${address}"
```
### Response
### 200 (Success)
```jsonc
{
  "amount": {
    "expected": 1,
    "covered": {
      "total": 200000000,
      "unlocked": 200000000,
      "locked": 0
    }
  },
  "complete": true,
  "description": "Donation to kernal",
  "created_at": "2021-07-11T19:04:24.574583Z",
  "transactions": [
    {
      "amount": 200000000,
      "confirmations": 4799,
      "double_spend_seen": false,
      "fee": 9200000,
      "height": 2402648,
      "timestamp": "2021-07-11T19:19:05Z",
      "tx_hash": "0c9a7b40b15596fa9a06ba32463a19d781c075120bb59ab5e4ed2a97ab3b7f33",
      "unlock_time": 0
    }
  ]
}
```

## Transfer
### Request
```sh
curl -s -X POST "${endpoint}/transfer" \
	-H 'Content-Type: application/json' \
	-d '{"destinations": [{"amount": 1337000000, "address": "47stn..."}]}'
```
### Response
#### 200 (Success)
```jsonc
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

## Transfer tracking
### Request
```sh
curl -s -X GET "${endpoint}/transfer/${tx_hash}"
```
### Response
#### 200 (Success)
```jsonc
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
  "timestamp": "2021-07-18T11:37:50Z",
  "unlock_time": 10,
  "tx_hash": "cf448effb86f24f81476c0012a6636700488e13accd91f8f43302ae90fed25ce"
}
```

## Callback payload
```jsonc
{
  "amount": 200000000,
  "fee": 9200000,
  "description": "callback test",
  "tx_hash": "0c9a7b40b15596fa9a06ba32463a19d781c075120bb59ab5e4ed2a97ab3b7f33",
  "address": "82j31dfbz1GPF7SWpusNjDAaucbit2NBZTMKyLYvqEfyUfWbRALx2bDaHDvvnbxngh56XRvqCYazsQ5xfGSAGWnYMciZVbe",
  "confirmations": 3297,
  "unlock_time": 0,
  "height": 2402648,
  "timestamp": "2021-07-11T19:19:05Z",
  "double_spend_seen": false
}
```

### Usage
```
$ ./moneropayd -h
Usage of ./moneropayd:
  -bind="localhost:5000": Bind address:port for moneropayd
  -intervals="1m,5m,15m,30m,1h": Comma seperated list of callback intervals
  -postgres-database="moneropay": Name for PostgreSQL database
  -postgres-host="localhost": PostgreSQL database address
  -postgres-password="": Password for PostgreSQL database
  -postgres-port=5432: PostgreSQL database port
  -postgres-username="moneropay": Username for PostgreSQL database
  -rpc-address="http://localhost:18082/json_rpc": Wallet RPC server address
  -rpc-password="": Password for monero-wallet-rpc
  -rpc-username="": Username for monero-wallet-rpc
  -transfer-mixin=8: Number of outputs from the blockchain to mix with (0 means no mixing)
  -transfer-priority=0: Set a priority for transactions
  -transfer-unlock-time=10: Number of blocks before the monero can be spent (0 to not add a lock)
```
```sh
#!/bin/sh
export RPC_ADDRESS='http://localhost:18083/json_rpc'
export RPC_USERNAME='kernal'
export RPC_PASSWORD='s3cure'
export POSTGRES_PASSWORD='s3cure'
./moneropayd
```
