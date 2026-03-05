# PM2 - Process Manager 2

A production process manager for Go applications, inspired by the Node.js PM2.

## 描述

- 项目代码均有TRAE IDE生成
- 项目采用Go语言编写
- 项目主要实现进程管理、集群模式、日志管理、监控、启动脚本、进程状态、保存进程列表、零停机重启等功能

## Table of Contents

- [TODOLIST](#todolist)
- [Installation](#installation)
- [Usage](#usage)
- [Features](#features)
- [Contributing](#contributing)
- [License](#license)

## TODOLIST

- 进程管理
  - 启动进程
  - 停止进程
  - 重启进程
  - 删除进程
- 集群模式
  - 启动集群
  - 停止集群
  - 重启集群
  - 删除集群
- 日志管理
  - 查看日志
  - 旋转日志
- 监控
  - 实时监控
  - 历史监控
- 启动脚本
  - 生成启动脚本
  - 配置启动脚本
- 进程状态
  - 查看进程状态
  - 重启进程状态
- 保存进程列表
  - 保存当前进程列表
  - 从保存的列表启动进程
- 零停机重启
  - 无中断重启应用程序

## Features

- **Process Management**: Start, stop, restart, and delete processes
- **Cluster Mode**: Run multiple instances of your application for better performance and reliability
- **Zero-downtime Reload**: Reload your application without losing any requests
- **Log Management**: Automatic log collection and rotation
- **Monitoring**: Real-time CPU and memory usage monitoring
- **Startup Script Generation**: Generate startup scripts for your applications
- **JSON Configuration**: Use JSON files to configure your applications
- **Name Uniqueness**: Ensure no duplicate application names

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/pm2.git

# Navigate to the project directory
cd pm2

# Build the binary
go build -o pm2 cmd/pm2/main.go

# Add to PATH (optional)
sudo mv pm2 /usr/local/bin/
```

## Usage

### Basic Commands

#### Start an Application

```bash
# Start a single instance
pm2 start ./app.js

# Start with a custom name
pm2 start ./app.js --name "my-app"

# Start in cluster mode with 4 instances
pm2 start ./app.js --instances 4

# Start from a JSON configuration file
pm2 start ecosystem.config.json
```

#### Stop an Application

```bash
# Stop by ID or name
pm2 stop <id|name>

# Stop all applications
pm2 stop all
```

#### Restart an Application

```bash
# Restart by ID or name
pm2 restart <id|name>

# Restart all applications
pm2 restart all
```

#### Delete an Application

```bash
# Delete by ID or name
pm2 delete <id|name>

# Delete all applications
pm2 delete all
```

#### List Applications

```bash
pm2 list
```

#### View Logs

```bash
# View logs for a specific application
pm2 logs <id|name>
```

#### Monitor Applications

```bash
pm2 monit
```

#### Generate Startup Script

```bash
pm2 startup
```

#### Save Process List

```bash
pm2 save
```

#### Reload with Zero Downtime

```bash
# Reload by ID or name
pm2 reload <id|name>

# Reload all applications
pm2 reload all
```

#### View Application Status

```bash
pm2 status <id|name>
```

#### View Application Details

```bash
pm2 describe <id|name>
```

## Configuration

### JSON Configuration File

You can use a JSON configuration file to define multiple applications. Here's an example `ecosystem.config.json`:

```json
{
  "apps": [
    {
      "name": "app",
      "script": "node",
      "args": ["app.js"],
      "instances": 1
    },
    {
      "name": "app2",
      "script": "node",
      "args": ["app2.js"],
      "instances": 2
    }
  ]
}
```

## Directory Structure

```
pm2/
├── cmd/
│   ├── pm2/
│   │   └── main.go        # PM2 CLI entry point
│   └── pm2-runtime/
│       └── main.go        # PM2 runtime entry point
├── internal/
│   ├── cluster/
│   │   ├── cluster.go     # Cluster implementation
│   │   └── manager.go     # Cluster manager
│   ├── command/
│   │   └── command.go     # Command implementations
│   ├── container/
│   │   └── runtime.go     # Container runtime
│   ├── log/
│   │   └── log.go         # Logging utilities
│   ├── monitor/
│   │   └── monitor.go     # Process monitoring
│   ├── process/
│   │   ├── manager.go     # Process manager
│   │   └── process.go     # Process implementation
│   └── startup/
│       └── startup.go     # Startup script generation
├── examples/              # Example applications
│   ├── README.md          # Examples documentation
│   ├── app.js             # Node.js example app
│   ├── app2.js            # Node.js example app 2
│   ├── ecosystem.config.json  # Example configuration file
│   ├── go-server.go       # Go HTTP server example
│   ├── node-server.js     # Node.js HTTP server example
│   └── python-server.py   # Python HTTP server example
├── pkg/                   # Public packages
├── test/                  # Test files
│   └── command_test.go    # Command tests
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── README.md              # This file
├── LICENSE                # MIT License
└── .gitignore             # Git ignore rules
```

## Examples

The `examples/` directory contains sample applications that demonstrate PM2's capabilities.

### Quick Start with Examples

```bash
# Navigate to examples directory
cd examples

# Start all applications from the ecosystem config
pm2 start ecosystem.config.json

# Check the status
pm2 list

# View logs
pm2 logs

# Stop all applications
pm2 stop all

# Delete all applications
pm2 delete all
```

### Individual Examples

#### Node.js Server

```bash
cd examples
pm2 start node-server.js
```

#### Go Server

```bash
cd examples
pm2 start go-server.go
```

#### Python Server

```bash
cd examples
pm2 start python-server.py
```

For more detailed examples, see the [examples/README.md](examples/README.md) file.

## License

MIT
