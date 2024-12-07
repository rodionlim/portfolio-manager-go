# Portfolio Valuation Tool

An application to value stocks, fx, commodities, cash, bonds and cryptocurrencies in your personal portfolio.

## Features

- Value assets based on current market prices
- Store portfolio data in a JSON file for easy access and manipulation
- Store portfolio data in leveldb for persistence
- Calculate total portfolio value based on current prices
- Display detailed information for individual assets

## Installation

1. Install Go version <b>1.23.4</b> or higher.
2. Clone the repository to your local machine.
3. Run `make` to build and install the application
4. Run the `portfolio-manager` binary to start the application

## Quickstart

Start the application

```sh
make run
```

Build the application

```sh
make
```

Wipe the entire database

```sh
make clean-db
```

## Key Features

1. Add assets to your portfolio using the `/blotter/trade` POST api.
2. View all assets in your portfolio and their total value using the `/portfolio/positions` GET api.
3. View detailed information for a specific asset using the `/asset` GET api.
4. Remove assets from your portfolio using the `/blotter/trade` DELETE api.

## Project Structure

```
portfolio-manager/
├── cmd/
│   └── portfolio/
│       └── main.go
├── internal/
│   ├── blotter/
│   ├── config/
│   ├── dal/
│   ├── portfolio/
│   └── server/
├── pkg/
│   ├── common/
│   ├── event/
│   ├── logging/
│   └── types/
├── .gitignore
├── go.mod
└── README.md
```

## Sample Curl Commands

### Add Asset

```sh
curl -X POST http://localhost:8080/blotter/trade \
    -H "Content-Type: application/json" \
    -d '{
        "symbol": "AAPL",
        "quantity": 10,
        "price": 150.00,
        "type": "buy"
    }'
```

### View Positions

```sh
curl -X GET http://localhost:8080/portfolio/positions
```

## Contributing

Contributions are always welcome! If you have any suggestions or find a bug, please open an issue or submit a pull request.

## License

This project is licensed under the MIT License - see the [license](./LICENSE) file for details.
