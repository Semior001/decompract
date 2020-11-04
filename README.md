# Decompract ![Go](https://github.com/Semior001/decompract/workflows/Go/badge.svg) [![Coverage Status](https://coveralls.io/repos/github/Semior001/decompract/badge.svg?branch=master)](https://coveralls.io/github/Semior001/decompract?branch=master)  [![godoc](https://godoc.org/github.com/semior001/decompract?status.svg)](https://godoc.org/github.com/Semior001/decompract) [![Go Report Card](https://goreportcard.com/badge/github.com/Semior001/decompract)](https://goreportcard.com/report/github.com/Semior001/decompract)

## Build and Deploy

### Environment variables
The application awaits next environment variables provided in .env file in the project folder:

| Environment       | Default  | Description                                                                                     | Example                                                        |
|-------------------|----------|-------------------------------------------------------------------------------------------------|----------------------------------------------------------------|
| DEBUG             | false    | Turn on debug mode                                                                              | true                                                           |
| SERVICE_URL       |          | URL to the backend service                                                                      | http://0.0.0.0:8080/                                           |
| SERVICE_PORT      | 8080     | Port of the backend servuce                                                                     | 8080                                                           |

### Run the application
```bash
docker-compose up -d
```

### Env file example

```.env
DEBUG=true
SERVICE_URL=http://0.0.0.0:8080/
SERVICE_PORT=8080
```

## Backend REST API

Several notes:
- All timestamps in RFC3339 format, like `2020-06-30T22:01:53+06:00`.
- All durations in RFC3339 format, like `1h30m5s`.
- Clocks should be represented in ISO 8601 format, like `15:04:05`.

### Errors format

#### Unauthorized
In case if the user requested a route without proper auth, the 401 status code will be returned with the `Unauthorized` body content.

#### General
Example:
```json
{
	"code"     : 0,
	"details"  : "failed to update event",
	"error"    : "event not found"
}
```

In case of bad client request error might have `null` value.

Supported error codes for client mapping:
```go
const (
	ErrInternal   ErrCode = 0 // any internal error
	ErrDecode     ErrCode = 1 // failed to unmarshal incoming request
	ErrBadRequest ErrCode = 2 // request contains incorrect data or doesn't contain data
)
```

### Client methods

