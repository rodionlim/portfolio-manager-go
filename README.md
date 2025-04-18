# Portfolio Valuation Tool

[![CI](https://github.com/rodionlim/portfolio-manager-go/actions/workflows/ci.yml/badge.svg)](https://github.com/rodionlim/portfolio-manager-go/actions/workflows/ci.yml)

An application to value equities, fx, commodities, cash, bonds (corps / gov), and cryptocurrencies in your personal portfolio.

## Features

- Value assets of different currencies based on current market prices
- Fetch market data based on free data sources (Yahoo finance, Google finance, dividends.sg, ilovessb.com, mas)
- Import / Export portfolio blotter data using CSV file for easy migration to other portfolio systems
- Allow users to supply their own custom dividends metadata
- Export ticker reference data in yaml format
- Autoclosing expired positions
- Store portfolio, reference, dividends and coupon data in leveldb for persistence
- Display detailed information for individual and aggregated assets
- OpenAPI compliant for easy integration with other systems
- UI for end users

## Installation

1. Install Go version <b>1.23.4</b> or higher.
2. Clone the repository to your local machine.
3. Run `make` to build and install the application
4. Run the `portfolio-manager` binary to start the application. Pass in config flag `-config custom-config.yaml`

### Proxmox VE Helper Scripts

For home-labbers, helpers scripts are exposed to allow easy installation of `portfolio-manager` in lxc containers within Proxmox VE.

```sh
bash -c "$(wget -qLO - https://github.com/rodionlim/portfolio-manager-go/raw/main/lxc/portfolio-manager.sh)"
```

## Quickstart

Start the application

```sh
make run # only start application backend
make run-full # if user wants to start with the UI, run this command instead
```

For Developers

```sh
make run # start backend
cd web/ui && npm run dev # start ui with hot reload on http://localhost:5173
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
├── lxc/
├── pkg/
│   ├── common/
│   ├── event/
│   ├── logging/
│   ├── mdata/
│   │   └── sources/
│   ├── rdata/
│   └── types/
├── templates/
│   └── blotter_import.csv # Sample for users to reference when trying to import blotter trades via csv
├── web/
├── .gitignore
├── go.mod
└── README.md
```

## UI

### Positions

Users can get an aggregated view of all their positions via the positions component in the user interface.

![Position Table](docs/Positions.png)

### Blotter Table

User can add, delete and update trades via the blotter component in the user interface.

![Blotter Table](docs/Blotter.png)

### Dividends Table

User can view dividends history of any given ticker at a granular level by ex-date

![Dividends Table](docs/Dividends.png)

### Settings

User can edit application wide settings, such as auto closing expired positions via the user interface

![Settings](docs/Settings.png)

## Backend API - Sample Curl Commands

All API calls are documented (OAS) under `http://localhost:8080/swagger/index.html`

### Add Asset to Blotter

```sh
curl -X POST http://localhost:8080/api/v1/blotter/trade \
    -H "Content-Type: application/json" \
    -d '{
        "ticker": "AAPL",
        "side": "buy",
        "broker": "DBS",
        "trader": "TraderA",
        "account": "CDP",
        "quantity": 10,
        "price": 150.00,
        "type": "buy",
        "tradeDate": "2024-12-09T00:00:00Z"
    }'
```

### Update Asset in Blotter

```sh
curl -X PUT http://localhost:8080/api/v1/blotter/trade \
    -H "Content-Type: application/json" \
    -d '{
        "ticker": "AAPL",
        "side": "buy",
        "broker": "DBS",
        "trader": "TraderA",
        "account": "CDP",
        "quantity": 10,
        "price": 200.00,
        "type": "buy",
        "tradeDate": "2024-12-09T00:00:00Z"
    }'
```

### Delete Assets from Blotter

```sh
curl -X DELETE http://localhost:8080/api/v1/blotter/trade \
    -H "Content-Type: application/json" \
    -d '["61570b49-2adb-4b99-be20-d14001e761a9"]'
```

### Delete all Assets from Blotter and Positions

```sh
curl -X DELETE http://localhost:8080/api/v1/blotter/trade/all
curl -X DELETE http://localhost:8080/api/v1/portfolio/positions
```

### Import Trades from CSV (for migrating into portfolio-manager)

```sh
curl -X POST http://localhost:8080/api/v1/blotter/import \
  -F "file=@templates/blotter_import.csv"
```

### Export Trades to a CSV (for migrating out of portfolio-manager)

```sh
curl -X GET http://localhost:8080/api/v1/blotter/export
```

### View Blotter Trades

```sh
curl -X GET http://localhost:8080/api/v1/blotter/trade
```

### View Positions

```sh
curl -X GET http://localhost:8080/api/v1/portfolio/positions
```

### Fetch Asset Prices

```sh
curl -X GET http://localhost:8080/api/v1/mdata/price/es3.si
curl -X GET http://localhost:8080/api/v1/mdata/price/temb
curl -X GET http://localhost:8080/api/v1/mdata/price/eth-usd
curl -X GET http://localhost:8080/api/v1/mdata/price/usd-sgd
```

### Fetch Dividends

```sh
# equity - refer to ticker reference for identifier
curl -X GET http://localhost:8080/api/v1/mdata/dividends/es3.si
curl -X GET http://localhost:8080/api/v1/mdata/dividends/aapl

# ssb - format SBMMMYY
curl -X GET http://localhost:8080/api/v1/mdata/dividends/sbjul24

# mas bill
curl -X GET http://localhost:8080/api/v1/mdata/dividends/bs24124z
```

### Store Custom Dividends

```sh
curl -X POST http://localhost:8080/api/v1/mdata/dividends/AAPL \
  -H "Content-Type: application/json" \
  -d '[
    {
      "ExDate": "2024-11-10",
      "Amount": 120.00,
      "AmountPerShare": 0.24,
      "Qty": 500
    },
    {
      "ExDate": "2024-08-09",
      "Amount": 115.00,
      "AmountPerShare": 0.23,
      "Qty": 500
    }
  ]'
```

### Fetch Portfolio Dividends

```sh
curl -X GET http://localhost:8080/api/v1/dividends/cjlu.si
```

### Fetch Reference Data

```sh
curl -X GET http://localhost:8080/api/v1/refdata
```

### Add Reference Data

```sh
curl -X POST http://localhost:8080/api/v1/refdata \
  -H "Content-Type: application/json" \
  -d '{
        "id": "ES3.SI",
        "name": "STI ETF",
        "underlying_ticker": "ES3.SI",
        "yahoo_ticker": "ES3.SI",
        "dividends_sg_ticker": "ES3",
        "asset_class": "eq",
        "asset_sub_class": "etf",
        "ccy": "SGD",
        "domicile": "SG"
      }'
```

### Delete Reference Data

```sh
curl -X DELETE http://localhost:8080/api/v1/refdata \
    -H "Content-Type: application/json" \
    -d '["ES3.SI"]'
```

### Export Reference Data

```sh
curl -X GET http://localhost:8080/api/v1/refdata/export
```

### Force a compute of dividends for a ticker across the entire blotter

```sh
curl -X POST http://localhost:8080/api/v1/dividends -H "Content-Type: application/json" -d '{"ticker": "ES3.SI"}'
```

### Autoclose expired positions

```sh
curl -X POST http://localhost:8080/api/v1/portfolio/cleanup
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

## Roadmap

1. Support non SGD dividends (Implemented)
2. Support MAS TBills (Implemented)
3. Support Crypto market data (Implemented)
4. Support FX market data (Implemented)
5. Support exporting/importing of leveldb for backup
6. Add UI component (Implemented)
7. Support lxc helper installation script (Implemented)
8. Support dividends viewer by ticker (Implemented)
9. Compute internal rate of return and store total portfolio value against price paid, with csv import/exports
10. Add UI charts for irr and portfolio value
11. Parallalize get position when fetching market data (Implemented)

## Contributing

Contributions are always welcome! If you have any suggestions or find a bug, please open an issue or submit a pull request.

## License

This project is licensed under the MIT License - see the [license](./LICENSE) file for details.
