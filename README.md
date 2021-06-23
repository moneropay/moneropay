# MoneroPay API (v1)
API for receiving, sending and tracking payments in Monero.
## Endpoints
| Method | URI                     | Input                                                           |
| :----: | ----------------------- | --------------------------------------------------------------- |
| `GET`  | /v1/balance             |                                                                 |
| `POST` | /v1/receive             | `'amount=123' 'description=desc' 'callback_url=http://merchant'` |
| `GET`  | /v1/receive/:subaddress |                                                                 |
| `POST` | /v1/transfer            | `{"destinations": [{"amount": 1337, "address": "47stn..."}]}`   |
| `GET`  | /v1/transfer/:tx_hash   |                                                                 |
| `GET`  | /v1/health              |                                                                 |

## Balance
### Request
```sh
curl -s -X GET "${endpoint}/v1/receive/${address}"
```
### Response
### 200 (Success)
```jsonc
{
	"total_balance": 157443303037455077,
	"unlocked_balance": 157360317826255077
}
```
## Receive
### Request
```sh
curl -s -X POST "${endpoint}/v1/receive"
	-d 'amount=123' # uint64 (required) - Amount to expect in XMR atomic units.
	-d 'description=Keep up the good work!' # string - The description for the order.
	-d 'callback_url=http://merchant' # string - Callback on incoming transfers.
```
### Response
### 200 (Success)
```jsonc
{
	"address": "85dd...", // Address to send payments to.
	"amount": 123,
	"description": "Keep up the good work!",
	"created_at": 1620165990 // Time of order was creation.
}
```

## Receipt tracking
### Request
```sh
curl -s -X GET "${endpoint}/v1/receive/${address}"
```
### Response
### 200 (Success)
```jsonc
{
	"amount": {
		"expected": 123,
		"covered": 100
	},
	"complete": false,
	"description": "Keep up the good work!",
	"created_at": 1620165990, // Time of order was creation.
	"transactions": [
		{
			"amount": 100,
			"confirmations": 8,
			"double_spend_seen": false,
			"fee": 21650200000,
			"height": 153624,
			"timestamp": 1620186597,
			"tx_hash": "c36258a276018c3a4bc1f195a7fb530f50cd63a4fa765fb7c6f7f49fc051762a",
			"unlock_time": 0
		}
	]
}
```

## Transfer
### Request
```sh
curl -s -X POST "${endpoint}/v1/transfer" -H 'Content-Type: application/json'
	-d '{"destinations": [{"amount": 1337, "address": "47stn..."}]}'
```
### Response
#### 200 (Success)
```jsonc
{
	"amount": 1337,
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
curl -s -X GET "${endpoint}/v1/transfer/${tx_hash}"
```
### Response
#### 200 (Success)
```jsonc
{
	"amount": 1337,
	"fee": 87438594,
	"state": "completed", // "pending", "completed" or "failed"
	"destinations": [
		{
			"amount": 1337,
			"address": "47stn..."
		}
	]
	"confirmations": 8,
	"double_spend_seen": false,
	"height": 153624,
	"timestamp": 1620186597,
	"unlock_time": 0,
	"tx_hash": "5ca34...",
}
```

## Health
### Request
```sh
curl -s -X GET "${endpoint}/v1/health"
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

### Usage
```
$ ./moneropayd -h
Usage of ./moneropayd:
  -bind string
        Bind address:port for moneropayd (default "localhost:5000")
  -postgres-database string
  	Name for PostgreSQL database
  -postgres-host string
  	PostgreSQL database address
  -postgres-password string
  	Password for PostgreSQL database
  -postgres-port uint
  	PostgreSQL database port
  -postgres-username string
  	Username for PostgreSQL database
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
