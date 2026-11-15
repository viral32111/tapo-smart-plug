# 🔌 Tapo Smart Plugs

[![CI](https://github.com/viral32111/tapo-smart-plug/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/viral32111/tapo-smart-plug/actions/workflows/ci.yml)
[![CodeQL](https://github.com/viral32111/tapo-smart-plug/actions/workflows/codeql.yml/badge.svg)](https://github.com/viral32111/tapo-smart-plug/actions/workflows/codeql.yml)
![GitHub tag (with filter)](https://img.shields.io/github/v/tag/viral32111/tapo-smart-plug?label=Latest)
![GitHub repository size](https://img.shields.io/github/repo-size/viral32111/tapo-smart-plug?label=Size)
![GitHub release downloads](https://img.shields.io/github/downloads/viral32111/tapo-smart-plug/total?label=Downloads)
![GitHub commit activity](https://img.shields.io/github/commit-activity/m/viral32111/tapo-smart-plug?label=Commits)

This is a Prometheus metrics exporter for [TP-Link's Tapo smart plugs](https://www.tp-link.com/uk/home-networking/smart-plug/) (primarily for the energy monitoring).

I use this for my [Tapo P110](https://www.tapo.com/uk/product/smart-plug/tapo-p110/)'s, [Tapo P110M](https://www.tp-link.com/uk/home-networking/smart-plug/tapo-p110m/)'s & [Tapo P304M](https://www.tp-link.com/uk/home-networking/smart-plug/tapo-p304m/)'s.

## 📥 Usage

Download the [latest release](https://github.com/viral32111/tapo-smart-plugs/releases/latest) for your platform. There are prebuilt binaries for Linux and Windows, covering both 32-bit and 64-bit architectures of x86 and ARM. **Linux builds are static with no libc dependency, thus work on both glibc & musl systems.**

### 🐳 Docker

There is a Docker image for Linux at [`ghcr.io/viral32111/tapo-smart-plugs`](https://github.com/users/viral32111/packages/container/package/tapo-smart-plugs). Please choose the appropriate tag for your use case:

| Tag | New features? | Bug fixes? | Breaking changes? | Notes |
| --- | --- | --- | --- | --- |
| `:1` | Yes | Yes | No | Recommended for short-term trials. |
| `:1.0` | No | Yes | No | Recommended for long-term stability. |
| `:1.0.0` | No | No | No | Not recommended unless dependent on a bug. |
| `:main` | Yes | Yes | Yes | Not recommended unless actively testing. |
| `:latest` | Yes | Yes | Yes | Not recommended unless actively testing. |

```bash
docker container run \
	--name tapo-smart-plugs \
	--mount type=bind,source="$PWD/tapo.toml",target=/etc/tapo.toml,readonly \
	--publish 5000:5000 \
	--rm \
	ghcr.io/viral32111/tapo-smart-plugs:1.0
```

## 📊 Metrics

All metrics are prefixed with `tapo_`.

* `tapo_status`: `1` if the smart plug is switched on, `0` if switched off.
* `tapo_uptime_seconds`: Number of seconds the smart plug has been switched on for.
* `tapo_current_milliwatts`: Current power consumption in milliwatts.
* `tapo_today_watt_hours`: Energy consumed today in watt-hours.
* `tapo_today_runtime_minutes`: Runtime today in minutes.
* `tapo_month_watt_hours`: Energy consumed this month in watt-hours.
* `tapo_month_runtime_minutes`: Runtime this month in minutes.

### Labels

* `id`: Device ID of the smart plug (or of the socket, for P304M child devices).
* `type`: Device type reported by the smart plug.
* `model`: Model of the smart plug (e.g., `P110`).
* `name`: Nickname set for the smart plug in the Tapo app.
* `firmware`: Firmware version of the smart plug.
* `hardware`: Hardware revision of the smart plug.
* `ssid`: Name of the Wi-Fi network the smart plug is connected to.
* `region`: Region configured on the smart plug.
* `language`: Language configured on the smart plug.
* `ip_address`: IP address of the smart plug.
* `mac_address`: MAC address of the smart plug.
* `slot`: Socket number on the power strip. Only applicable to the individual sockets of a P304M, empty otherwise.

## ⚙️ Configuration

The smart plugs to monitor, and optionally the account credentials, are registered in a TOML configuration file. Create one similar to the example below, then provide its path with `--config` (defaults to `/etc/tapo.toml`).

```toml
[log]
level = "info" # $TAPO_LOG_LEVEL

# The HTTP server for exposing metrics
[metrics]
listen = "127.0.0.1:5000" # $TAPO_HTTP_ADDRESS & $TAPO_HTTP_PORT
path = "/metrics" # $TAPO_METRICS_PATH
go_builtin = false # $TAPO_METRICS_GO_BUILTIN
interval = "10s" # $TAPO_METRICS_INTERVAL

# The TP-Link Tapo account credentials (used only for local authentication)
[account]
email = "john@example.com" # $TAPO_ACCOUNT_EMAIL
password = "P4ssw0rd!" # $TAPO_ACCOUNT_PASSWORD

# Add each smart plug to monitor like below.
# Account credentials can be overridden for each plug if needed (e.g. some plugs are owned by a different account).
[[plugs]]
name = "TV"
ip_address = "192.168.10.10"
# email = "alice@example.com"
# password = "P4ssw0rd@"

[[plugs]]
name = "Lamp"
ip_address = "192.168.10.15"

[[plugs]]
enabled = false # Defaults to true, set to false to prevent monitoring & exporting without removing it from the config
name = "Heater"
ip_address = "192.168.10.20"
```

Core configuration is also possible via environment variables for easy deployment with Docker, see [`.env.example`](/.env.example). These take priority over the TOML configuration file, each one overriding only its own part (e.g. `TAPO_HTTP_PORT` replaces just the port of `metrics.listen`). A variable that is set to an invalid value is an error, the exporter will refuse to start.

| Name | Type | Description | Default |
| --- | --- | --- | --- |
| `TAPO_LOG_LEVEL` | String | Minimum level of logs to print (`debug`, `info`, `warn`, `error`). Also configurable as `log.level` in the TOML file. | `info` |
| `TAPO_HTTP_ADDRESS` | String | IP address to listen for HTTP requests on. Also configurable as the host of `metrics.listen` in the TOML file. | `127.0.0.1` |
| `TAPO_HTTP_PORT` | Number | TCP port to listen for HTTP requests on (`1` to `65535`). Also configurable as the port of `metrics.listen` in the TOML file. | `5000` |
| `TAPO_METRICS_PATH` | String | HTTP path to expose Prometheus metrics on. Must start with `/`, must not end with `/` (unless it is just `/`), and must not contain whitespace. Also configurable as `metrics.path` in the TOML file. | `/metrics` |
| `TAPO_METRICS_GO_BUILTIN` | Boolean | Whether to include Go runtime & process metrics. Also configurable as `metrics.go_builtin` in the TOML file. | `false` |
| `TAPO_METRICS_INTERVAL` | Duration | Interval to poll each smart plug at, must be greater than zero (e.g. `10s`). Also configurable as `metrics.interval` in the TOML file. | `10s` |
| `TAPO_ACCOUNT_EMAIL` | String | Email address of the Tapo account. Required, unless `account.email` is set in the TOML file. | |
| `TAPO_ACCOUNT_PASSWORD` | String | Password of the Tapo account. Required, unless `account.password` is set in the TOML file. | |

### 🚩 Flags

| Flag | Description |
| --- | --- |
| `-v`, `--version` | Print the version & exit. |
| `-u`, `--usage` | Show the usage message & exit. `-h` and `--help` are aliases. |
| `-c`, `--config <path>` | Path to the TOML configuration file. Defaults to `/etc/tapo.toml`. |

## 🧰 Development

1. Copy the [`.env.example`](/.env.example) file to `.env` and populate accordingly.
2. Run the application with `go run ./cmd/tapo/main.go --config <path>`.

## 🛠️ Building

Run the relevant build command below, assuming you're on a Linux system.

### Linux 64-bit

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -v -ldflags="-s -w -X main.VERSION=0.0.0" -o tapo-smart-plugs-v0.0.0-linux-amd64 ./cmd/tapo/main.go
```

### Windows 64-bit

```bash
GOOS=windows GOARCH=amd64 \
	go build -v -ldflags="-s -w -X main.VERSION=0.0.0" -o tapo-smart-plugs-v0.0.0-windows-amd64.exe ./cmd/tapo/main.go
```

## ⚖️ License

Copyright (C) 2026 [viral32111](https://viral32111.com).

> This program is free software: you can redistribute it and/or modify
> it under the terms of the GNU Affero General Public License as
> published by the Free Software Foundation, either version 3 of the
> License, or (at your option) any later version.

> This program is distributed in the hope that it will be useful,
> but WITHOUT ANY WARRANTY; without even the implied warranty of
> MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
> GNU Affero General Public License for more details.

> You should have received a copy of the GNU Affero General Public License
> along with this program. If not, see https://www.gnu.org/licenses.
