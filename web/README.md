# Portfolio Manager Web

This project contains the web interface for the Portfolio Manager application. It includes a React application built with Mantine, Vite and a Go server to serve the built application.

## Prerequisites

- Node.js >= 20
- npm >= 7

## Setup

### Install Node.js and npm

Make sure you have Node.js and npm installed. You can download them from [nodejs.org](https://nodejs.org/).

### Install Dependencies

Navigate to the `web/ui` directory and install the npm dependencies:

```sh
cd web/ui
npm install
```

## Development

### Run the React Application

To start the React development server, run:

```sh
npm run dev
```

This will start the development server on `http://localhost:5173`.

### Build the React Application

To build the React application for production, run:

```sh
npm run build
```

This will create a `build` directory inside the `web/ui` directory with the production build of the application.

### Testing

Testing is done via `vitest`.

To run specific tests

```sh
npm run test ReportsTable
```
