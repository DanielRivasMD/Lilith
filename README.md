# Lilith, master of daemons

[![License](https://img.shields.io/badge/license-GPLv3-blue.svg)](LICENSE)
[![Documentation](https://godoc.org/github.com/DanielRivasMD/Lilith?status.svg)](http://godoc.org/github.com/DanielRivasMD/Lilith)
[![Go Report Card](https://goreportcard.com/badge/github.com/DanielRivasMD/Lilith)](https://goreportcard.com/report/github.com/DanielRivasMD/Lilith)
[![Release](https://img.shields.io/github/release/DanielRivasMD/Lilith.svg?label=Release)](https://github.com/DanielRivasMD/Lilith/releases)


## Overview
Go-based CLI for orchestrating, monitoring, and controlling background processes with precision
It is built for developers & operators who need reliable, fine‑grained control over concurrent jobs without the noise

## Features

### Core capabilities

- **Process orchestration**: Spawn & manage background processes with structured metadata and predictable lifecycle control
- **Grouping and workflows**: Assign related processes to groups for coordinated start, stop, & teardown
- **Signal control**: Pause, resume, & terminate processes via standard system signals
- **Status and history**: Inspect live state, invocation history, exit codes, & runtimes at a glance
- **File watching**: Trigger scripts or tasks when monitored paths change, with debouncing & clean restarts

### Use cases

- **Development loops**: Run linters, test suites, & rebuilds in parallel with clear visibility & control.
- **Automation pipelines**: Chain scripts & long‑running tasks with grouping & graceful shutdowns.
- **Ops tooling**: Keep lightweight daemons in check, audit their status, & enforce consistent process behavior.

### Design goals

- **Predictable**: Clear, composable primitives for starting, grouping, & signaling processes.
- **Observable**: First‑class status, logs, & history so you can see what’s running & why.
- **Minimal**: No hidden magic; sane defaults with explicit configuration when you need it.

## Quickstart
```
```

## Installation

### **Language-Specific**
| Language   | Command                                                                 |
|------------|-------------------------------------------------------------------------|
| **Go**     | `go install github.com/DanielRivasMD/Lilith@latest`                  |

### **Pre-built Binaries**
Download from [Releases](https://github.com/DanielRivasMD/Lilith/releases).

## Usage

```
```

| Command     | Description                            |
|-------------|----------------------------------------|
| `invoke`    | Start a new daemon                     |
| `freeze`    | Pause a running daemon                 |
| `rekindle`  | Resurrect a paused or limbo daemon     |
| `slay`      | Stop and clean up daemon processes     |
| `tally`     | List all active daemons                |
| `summon`    | View logs of specific daemon(s)        |
| `help`      | Display help for any command           |


## Example
```
```

## Configuration
<!-- TODO: add instructions for installing local directory & mock config-workflow file -->
<!-- TODO: explain how logic works -->


## Development

Build from source
```
git clone https://github.com/DanielRivasMD/Lilith
cd Lilith
```

## Language-Specific Setup

| Language | Dev Dependencies | Hot Reload           |
|----------|------------------|----------------------|
| Go       | `go >= 1.21`     | `air` (live reload)  |

## License
Copyright (c) 2025

See the [LICENSE](LICENSE) file for license details.
