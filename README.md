# Portfolio Valuation Tool

A command line application to value stocks and cryptocurrencies in your personal portfolio.

## Features

- Valuate stocks and cryptocurrencies based on current market prices
- Store portfolio data in a JSON file for easy access and manipulation
- Store portfolio data in leveldb for persistence
- Calculate total portfolio value based on current prices
- Display detailed information for individual assets

## Installation

1. Install Go version 1.16 or higher.
2. Clone the repository to your local machine.
3. Install the required dependencies by running go get.
4. Run `go build` to build the application.
5. Run the executable file to launch the application.

## Usage

1. Add assets to your portfolio using the add command.
2. View all assets in your portfolio using the view command.
3. View detailed information for a specific asset using the info command.
4. Remove assets from your portfolio using the remove command.
5. Calculate the total value of your portfolio using the total command.

## Project Structure

```
portfolio-manager/
├── cmd/
│ └── portfolio/
│ ├── main.go
├── internal/
│ └── portfolio/
│   ├── portfolio.go
│   ├── asset.go
│   ├── add.go
│   ├── asset.go
│   ├── info.go
│   ├── remove.go
│   ├── total.go
│   ├── view.go
│   └── json.go│
│ └── config/
│   ├── config.go
│ └── dal/
│   ├── levledb.go
├── .gitignore
├── go.mod
└── README.md
```

## Contributing

Contributions are always welcome! If you have any suggestions or find a bug, please open an issue or submit a pull request.

## License

This project is licensed under the MIT License - see the [license](./LICENSE) file for details.
