# Portfolio Valuation Tool

An application to value equities, fx, commodities, cash, bonds and cryptocurrencies in your personal portfolio.

## Features

- Value assets based on current market prices
- Fetch market data based on free data sources (Yahoo finance, Google finance, dividends.sg) with local cache
- Output portfolio data in a CSV file for easy access and manipulation
- Import / Export portfolio data from CSV file for easy migration to other portfolio maangers
- Store portfolio, reference and dividends data in leveldb for persistence
- Calculate total portfolio value based on current prices
- Display detailed information for individual assets
- OpenAPI compliant for easy integration with other systems

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

## Project Structure

```
portfolio-manager/
├── cmd/
│   └── portfolio/
│       └── main.go
├── docs/
│   └── swagger.json
├── internal/
│   ├── blotter/
│   ├── config/
│   ├── dal/
│   ├── dividends/
│   ├── mocks/
│   ├── portfolio/
│   └── server/
├── pkg/
│   ├── common/
│   ├── event/
│   ├── logging/
│   ├── mdata/
│   │   └── sources/
│   ├── rdata/
│   └── types/
├── .gitignore
├── go.mod
└── README.md
```

## Sample Curl Commands

All API calls are documented (OAS) under `http://localhost:8080/swagger/index.html`

### Add Asset

```sh
curl -X POST http://localhost:8080/blotter/trade \
    -H "Content-Type: application/json" \
    -d '{
        "ticker": "AAPL",
        "side": "buy",
        "broker": "dbs",
        "trader": "traderA",
        "quantity": 10,
        "price": 150.00,
        "type": "buy",
        "tradeDate": "2024-12-09T00:00:00Z"
    }'
```

### Import Trades from CSV (for migrating into portfolio-manager)

```sh
curl -X POST http://localhost:8080/blotter/import \
  -F "file=@templates/blotter_import.csv"
```

### Export Trades to a CSV (for migrating out of portfolio-manager)

```sh
curl -X GET http://localhost:8080/blotter/export
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
curl -X GET http://localhost:8080/mdata/dividend/es3.si
```

### Fetch Reference Data

```sh
curl -X GET http://localhost:8080/refdata
```

### Force a compute of dividends for a ticker across the entire blotter

```sh
curl -X POST http://localhost:8080/dividends -H "Content-Type: application/json" -d '{"ticker": "ES3.SI"}'
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
divWitholdingTaxSG: 0
divWitholdingTaxUS: 0.3
divWitholdingTaxHK: 0
divWitholdingTaxIE: 0.15
```

## Contributing

Contributions are always welcome! If you have any suggestions or find a bug, please open an issue or submit a pull request.

## License

This project is licensed under the MIT License - see the [license](./LICENSE) file for details.
