# Lilith, master of daemons

[![License](https://img.shields.io/badge/license-GPLv3-blue.svg)](LICENSE)
[![Documentation](https://godoc.org/github.com/DanielRivasMD/Lilith?status.svg)](http://godoc.org/github.com/DanielRivasMD/Lilith)
[![Go Report Card](https://goreportcard.com/badge/github.com/DanielRivasMD/Lilith)](https://goreportcard.com/report/github.com/DanielRivasMD/Lilith)
[![Release](https://img.shields.io/github/release/DanielRivasMD/Lilith.svg?label=Release)](https://github.com/DanielRivasMD/Lilith/releases)

## Overview

Minimalist CLI for orchestrating, monitoring & controlling background processes
that watch filesystem changes and execute scripts

`lilith` spawns persistent watcher daemons, keeps structured metadata and logs,
tracks process state, and lets you pause, resume, or terminate them

## Technical Architecture

Lilith is a Go‑based CLI that separates command‑line orchestration from
background workers, persists per‑daemon state on disk, and uses standard Unix
signals for process control

### Core Framework

- Built with **Cobra** for command definitions & **Viper** for loading TOML
  workflows
- Each daemon is a **watchexec** child process, launched and detached by the
  Lilith launcher
- Metadata (PID, watch directory, script, group, invocation time) is stored as
  JSON
- Process signals (`SIGSTOP`, `SIGCONT`, `SIGTERM`) are used to freeze,
  revive, and slay daemons

### Logic Schematic

    ┌────────────────┐
    │ lilith genesis │ → creates directories + example config
    └───────┬────────┘
            │
            ▼
    ┌─────────────────────────────────┐
    │ lilith invoke <workflow>        │
    │   - loads config or uses flags  │
    │   - checks for duplicate        │
    │   - spawns watcher (detached)   │
    │   - writes metadata             │
    └───────┬─────────────────────────┘
            │
            ▼
    ┌─────────────────────────────────────────┐
    │ watchexec --watch <dir> -- script       │ (background worker)
    │   - logs stdout/stderr to ~/.lilith/log │
    │   - runs on every filesystem change     │
    └───────┬─────────────────────────────────┘
            │
            ▼
    ┌───────────────────────────────────┐
    │ lilith tally                      │ → lists all daemons with status
    │ lilith summon <daemon> [--follow] │ → view / tail log
    │ lilith freeze <daemon>            │ → SIGSTOP
    │ lilith revive <daemon>            │ → SIGCONT (or respawn if dead)
    │ lilith slay <daemon>              │ → SIGTERM + remove metadata/log
    └───────────────────────────────────┘

### Storage Layout (~/.lilith/)

    ~/.lilith/
    ├─ config/   # workflow definitions (*.toml)
    ├─ log/      # logs for each daemon (*.log)
    └─ daemon/   # metadata for each running daemon (*.json)

### Workflow Configuration Example

    # ~/.lilith/config/forge.toml

    [workflows.helix]
    watch  = "~/src/helix"
    script = "~/.lilith/forge/helix.sh"
    daemon = "helix-watcher"      # optional
    group  = "forge"              # optional
    log    = "helix"              # optional

## Installation

### Language-Specific

    Go:  go install github.com/DanielRivasMD/Lilith@latest

## License

Copyright (c) 2025

See the [LICENSE](LICENSE) file for license details.
