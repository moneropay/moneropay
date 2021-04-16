# MoneroPay

## `moneropayd` - Monero payment API

### Endpoints
| Method | URI                | Data                                          |
| :----: | ------------------ | --------------------------------------------- |
| `GET`  | /v1/balance/       |                                               |
| `POST` | /v1/address/       |                                               |
| `GET`  | /v1/address/:index |                                               |
| `POST` | /v1/transfer/      | `[{"amount": 0.1337, "address": "47stn..."}]` |
| `GET`  | /v1/ping/          |                                               |

### Responses
#### GET /v1/balance/
```json
{
  "total_balance": 0.00204076,
  "unlocked_balance": 0.00204076
}
```

#### POST /v1/address/
```json
{
  "address": "88HjkMjkPnRXnxkwQJPGEj7Jd6TfAqSn5bAiJNrB6J3PYsoTCW4dX6DHRknxizMaRwZ27WPmtYMfwc9RaHrBVQfSHDJn2R7",
  "index": 22
}
```

#### GET /v1/address/:index
```json
{
  "transfers": [
    {
      "address": "85ddLJ1qMmsWwtC9gmFi1S5uuPWKT5puQisQitTqNpwBUgw9gEYTmszLfQXygNUhrSGy8fzq3CEEXFkNdKHFbHWkFE3JS1s",
      "amount": 0.001,
      "confirmations": 26140,
      "double_spend_seen": false,
      "fee": 1.097e-05,
      "timestamp": 1614871910,
      "txid": "6685406e07564bb4bf90553e5a30c0ef2afab4e13393b9f773964fdcf13a39c2",
      "unlock_time": 0
    }
  ]
}
```

#### POST /v1/transfer/
```json
{
  "amount": 0.0001312,
  "data": [
    {
      "amount": 0.0001212,
      "address": "47stnbyV5rqaCuinwxmWHH2qPEQj5PCbkipBYoNvkhxDCkFk9Qo4ijvNF9EfTmG1TTcoCxUuk97GrfnUQReVNYYT6SCqUA8"
    },
    {
      "amount": 1e-05,
      "address": "46VGoe3bKWTNuJdwNjjr6oGHLVtV1c9QpXFP9M2P22bbZNU7aGmtuLe6PEDRAeoc3L7pSjfRHMmqpSF5M59eWemEQ2kwYuw"
    }
  ],
  "fee": 1.869e-05,
  "tx_hash": "5ca343fbebde6841b9c653c2ea03d08ab22113fcda1cfa539d3feccbf84321a9"
}
```

#### GET /v1/ping/
```json
"pong"
```

### Usage
```
$ ./moneropayd -h
Usage of ./moneropayd:
  -bind string
        Bind address:port for moneropayd (default "localhost:5000")
  -rpc-address string
        Wallet RPC server address (default "http://localhost:18082/json_rpc")
  -rpc-password string
        Password for monero-wallet-rpc
  -rpc-username string
        Username for monero-wallet-rpc
  -transfer-mixin uint
        Number of outputs from the blockchain to mix with (0 means no mixing) (default 8)
  -transfer-priority uint
        Set a priority for transactions
  -transfer-unlock-time uint
        Number of blocks before the monero can be spent (0 to not add a lock) (default 10)
```
