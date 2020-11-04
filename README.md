# Decompract ![Go](https://github.com/Semior001/decompract/workflows/Go/badge.svg) [![codecov](https://codecov.io/gh/Semior001/decompract/branch/master/graph/badge.svg?token=IW8CU4ZDG6)](https://codecov.io/gh/Semior001/decompract) 

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
EMAIL=e.duskaliev@innopolis.university
PASSWORD=test

SECRET=hO]rqN2cQ|r/oOWTS~*Q3=@w-zd<c"
DB_TEST=postgres://attc:attcpwd@localhost:5432/attc?sslmode=disable

# for postgres service
PG_USER=attc
PG_PWD=attcpwd
PG_DB=attc
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

#### Auth methods
- `GET /auth/local/login` - authenticate and get JWT token. The token will be saved in secure cookies. 
  - Body:
    ```json
    {
        "user": "e.duskaliev@innopolis.university",
        "passwd": "verystrongpassword"
    }
    ```
  - Response (example, shrinked for the sake of simplicity; next examples will be also shrinked): 
    - Headers:
        ```text
        Set-Cookie: JWT=json.web.token; Path=/; Max-Age=720000; HttpOnly
        ```
    - Body (avatar is not used in the application, it is provided by the auth library):
        ```json
        {
          "name": "e.duskaliev@innopolis.university",
          "id": "local_7f48448389aa065af161c3215237acef139e4ecf",
          "picture": "http://0.0.0.0:8080/avatar/", 
          "email": "e.duskaliev@innopolis.university",
          "attrs": {
            "privileges": [
              "read_users",
              "edit_users",
              "list_users",
              "add_users"
            ]
          }
        }
        ```

- `GET /auth/local/user` - get information about the logged in user. Requires authed user.
  - Body: `empty`
  - Response: 
    ```json
    {
      "name": "e.duskaliev@innopolis.university",
      "id": "local_7f48448389aa065af161c3215237acef139e4ecf",
      "picture": "http://0.0.0.0:8080/avatar/", 
      "email": "e.duskaliev@innopolis.university",
      "attrs": {
        "privileges": [
          "read_users",
          "edit_users",
          "list_users",
          "add_users"
        ]
      }
    }
    ```

- `GET /auth/local/logout` - logout from the app, this will remove the JWT token from cookies.
  - Body: `empty`
  - Response:
    - Headers:
      ```text
      Set-Cookie: JWT=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Max-Age=0
      ```
    - Body: `empty`
