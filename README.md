# Companies House Technical Exercise

This is my implementation of the companies house technichal exercise.
It is written in Go, using logrus for logging and gorilla mux for the router.
It is backend independent an can use any backend that implements the backend.DataSource 
interface. The backend implemented is mongodb.

## Requirements 

- mongodb instance 
- Go compiler 

## Build & Run

To build the micro service use the command 
```
$ go build server.go
```

To run the server use the command 
```
$ go run server.go 
```

### Args

The micro service can be configured with command line arguments. Namely:

- -p - Set the port to start the micro service on
- -v - Set the logging level to debug

## File structure

```
.
├── server.go           - Handles args, starts server
├── backend             - backend implementations
│   ├── datasource.go   - Interface definition for backend
│   ├── model.go        - Model structure definitions
│   └── mongo.go        - MongoDB implementation of the ServiceDataSource interface 
└── service             - Main service package
    ├── gameservice     - GaneService package
    │   ├── handler.go  - GameService http.Handler
    │   └── handler_test.go
    ├── service.go      - Main Service http.Handler
    └── service_test.go

```