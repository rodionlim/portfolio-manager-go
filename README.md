# Portfolio Valuation Tool

An application to value equities, fx, commodities, cash, bonds and cryptocurrencies in your personal portfolio.

## Features

- Value assets based on current market prices
- Fetch market data based on free data sources (Yahoo finance, Google finance, dividends.sg)
- Output portfolio data in a CSV file for easy access and manipulation
- Import portfolio data from CSV file for easy migration into portfolio-manager-go
- Export portfolio data to CSV file for easy migration to another portfolio manager
- Store portfolio data and reference data in leveldb for persistence
- Calculate total portfolio value based on current prices
- Display detailed information for individual assets

## Installation

1. Install Go version <b>1.23.4</b> or higher.
2. Clone the repository to your local machine.
3. Run `make` to build and install the application
4. Run the `portfolio-manager` binary to start the application. Pass in config flag `-config custom-config.yaml`

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

Tests

```sh
make test # unit tests
make test-integration # integration tests
```

## Endpoints

1. Add assets to your portfolio using the `/blotter/trade` POST api.
2. View all assets in your portfolio and their total value using the `/portfolio/positions` GET api.
3. View detailed information for a specific asset using the `/asset` GET api.
4. Remove assets from your portfolio using the `/blotter/trade` DELETE api.
5. Fetch market data using the `/mdata/ticker` GET api.

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
│   ├── mdata/
│   │   └── sources/
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
        "ticker": "AAPL",
        "side": "buy",
        "assetClass": "eq",
        "broker": "dbs",
        "trader": "traderA",
        "quantity": 10,
        "price": 150.00,
        "type": "buy",
        "tradeDate": "2024-12-09T00:00:00Z"
    }'
```

### Import Trades from CSV

```sh
curl -X POST http://localhost:8080/blotter/import \
  -F "file=@templates/blotter_import.csv"
```

### View Positions

```sh
curl -X GET http://localhost:8080/portfolio/positions
```

### Fetch Stock Prices

```sh
curl -X GET http://localhost:8080/mdata/price/es3.si
```

### Fetch Dividends

```sh
curl -X GET http://localhost:8080/mdata/dividend/es3
```

### Fetch Reference Data

```sh
curl -X GET http://localhost:8080/refdata
```

## Configurations

Sample configurations

```yaml
verboseLogging: true
logFilePath: ./portfolio-manager.log
host: localhost
port: 8080
db: leveldb
dbPath: ./portfolio-manager.db
refDataSeedPath: "./seed/refdata.yaml"
```

## Contributing

Contributions are always welcome! If you have any suggestions or find a bug, please open an issue or submit a pull request.

## License

This project is licensed under the MIT License - see the [license](./LICENSE) file for details.
