# PM2 Examples

This directory contains example applications that you can run with PM2.

## Node.js Server

A simple Node.js HTTP server that listens on port 3000 by default.

### Usage

```bash
# Start the server with PM2
pm2 start node-server.js

# Check the status
pm2 list

# Stop the server
pm2 stop node-server.js
```

## Go Server

A simple Go HTTP server that listens on port 8080 by default.

### Usage

```bash
# Build the Go server
go build go-server.go

# Start the server with PM2
pm2 start ./go-server

# Check the status
pm2 list

# Stop the server
pm2 stop go-server
```

## Python Server

A simple Python HTTP server that listens on port 8000 by default.

### Usage

```bash
# Start the server with PM2
pm2 start python-server.py

# Check the status
pm2 list

# Stop the server
pm2 stop python-server.py
```

## Using JSON Configuration

You can also use a JSON configuration file to start multiple applications at once.

### Example ecosystem.config.json

```json
{
  "apps": [
    {
      "name": "node-server",
      "script": "node-server.js",
      "instances": 1,
      "env": {
        "PORT": 3000
      }
    },
    {
      "name": "go-server",
      "script": "./go-server",
      "instances": 1,
      "env": {
        "PORT": 8080
      }
    },
    {
      "name": "python-server",
      "script": "python-server.py",
      "instances": 1,
      "env": {
        "PORT": 8000
      }
    }
  ]
}
```

### Usage

```bash
# Start all applications from the config file
pm2 start ecosystem.config.json

# Check the status
pm2 list

# Stop all applications
pm2 stop all
```
