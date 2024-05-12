
# Problem 3: Transaction Broadcasting and Monitoring Client 

## TechStack
All are written in Go, here is the library I'm using
- Echo
- Viper
- Zap
- Validator V10
## Table of Contents

- [How to run](#how-to-run)
- [API Endpoints](#api-endpoints)
  - [Liveness Check](#liveness-check)
  - [Broadcast and check status](#broadcast-and-waiting-for-transaction-status)
  - [Check status](#checking-transaction-status)
  - [Response](#response)
- [Configuration](#configuration)

## How to run

To get started, follow these steps:

1. **Clone the Repository:**
   ```
   git clone https://github.com/BoomNooB/post-get-check-api.git
   cd post-get-check-api
   ```
   

4. **Running a makefile:**\
   Simply just type this to your terminal after `cd` from step one
   `make run`
   then it will do a couple of thing here
   - `go mod tidy` for download and install dependencies  
   - `go run ./cmd/main.go` for running the program

   The server will start at `http://localhost:8080` (you can change the port as you wish in `config/config.yaml`)

## API Endpoints

### Liveness Check

Check if the API is running.

- **URL:** `/health`
- **Method:** `GET`
- **Response:**
  - Status Code: 200 OK
  - Body:
```
"Service is healthy!"
```
### Broadcast and waiting for transaction status

This will broadcast for transaction and get transaction hash
then it will be checking status and return status of that transaction from transaction hash
and if status is pending it will retry until reach point that we've set in `config.yaml` file

- **URL:** `/broadcast/`
- **Method:** `POST`
- **Request Header:** 
	- `X-Request-ID` it can be any string value using for trace in case of not happy case happened
- **Request Body:** 
	- All field are REQUIRED 
	- `timestamp` have to be `10-13` digits (UNIX epoch timestamps format)
```
{
  "symbol":"VTX", //must be string 
  "timestamp":1234567890213, //must be uint64
  "price":1234 //must be uint64
}
```

### Checking transaction status

This will perform checking status of that transaction only one time

- **URL:** `/check/pending/`
- **Method:** `GET`
- **Request Header:** 
	- `X-Request-ID` it can be any string value using for trace in case of not happy case happened
- **Request Body:**
```
  {
    "tx_hash": "781e812ac6b5542d320b3da916b13a431cacbe01caee19476efb3641965f2877"
    // it's a hash that get from broadcasting in case of it's still pending after retry
  }
```


### Response
There's many response that can be happen but the happy one will look like this with status: `200` `ok`
```
{
  "msg": "Status checking success",
  "tx_status": "CONFIRMED",
  "tx_hash": "2fe537287be3525eedf1c4fc3df82c16b53800c745db8ae0f07d7ea485522c4d"
} 
```
the `msg` is message that telling what happend which is can be

- `Cannot bind request`  
	- status: `400` `Bad Request`
	- will occur when program cannot bind request mostly happen when the request body in wrong format
- `Request header is invalid, X-Request-ID is REQUIRED`
	- status: `400` `Bad Request`
	- will occur when you not send `X-Request-ID` in header
- `Request body is invalid`
	- status: `400` `Bad Request`
	- will occur when you not sending require field in body as mentioned above 
- `Request body is invalid`
	- status: `400` `Bad Request`
	- will occur when you not sending require field in body as mentioned above 
- `Cannot broadcast and check for transactions`
	- status: `500` `Internal Server Error`
	- will occur when program is crash or cannot perform operation 
- `Cannot check for transactions`
	- status: `500` `Internal Server Error`
	- will occur when program is crash or cannot perform operation in `/check/pending/`
- `After retry n times, status are still PENDING, please check via /check/pending/"`
	- status: `200` `OK`
	- will occur when program is trying to check status after n time and status are still pending



## Configuration

you can simply edit some parameter as following in `config/config.yaml` file to achieve your wish
```
app:
	# port for starting program
	port: :8080 
	
logger:
	# this can be dev or prod
	env: prod 

# time for graceful shutdown in case of exit program while working
context_timeout_graceful: 10s 

# timeout of http client that call outside api
http_client_timeout: 10s 

api_path:
	# path of health check
	health-check: /health 

	# path to get tx_hash
	post-txn: https://some.app/broadcast/
	
	# path to check tx_status with tx_hash
	get-txn: https://some.app.app/check/
	
	# path for the program to perform broadcast and check
	broadcast-ext-txn-path: /broadcast/

	# path for the program to perform check
	check-ext-txn-pending: /check/pending/


retry_for_check:

	# number of time to retry checking status in broadcast and checking path
	retry_times: 6

	# delay of each retry before starting new one
	retry_repeat_delay: 10s
```

---
## What can be improve
1. In case of heavy load for API, we might use Kafka as a messaging queue to handle before user reach this API
2. Each request should store in database or cache in case of want to reconcile data
3. In the program `logger` should be wrapped before use to show `X-Request-ID` for each, now I've to put `zap.Any` for each time called log to add `X-Request-ID`
4. Unit test should be implemented testing service file