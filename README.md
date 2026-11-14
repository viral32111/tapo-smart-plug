# Tapo Smart Plugs

[![CI](https://github.com/viral32111/tapo-smart-plug/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/viral32111/tapo-smart-plug/actions/workflows/ci.yml)
[![CodeQL](https://github.com/viral32111/tapo-smart-plug/actions/workflows/codeql.yml/badge.svg)](https://github.com/viral32111/tapo-smart-plug/actions/workflows/codeql.yml)
![GitHub tag (with filter)](https://img.shields.io/github/v/tag/viral32111/tapo-smart-plug?label=Latest)
![GitHub repository size](https://img.shields.io/github/repo-size/viral32111/tapo-smart-plug?label=Size)
![GitHub release downloads](https://img.shields.io/github/downloads/viral32111/tapo-smart-plug/total?label=Downloads)
![GitHub commit activity](https://img.shields.io/github/commit-activity/m/viral32111/tapo-smart-plug?label=Commits)

This is a Prometheus metrics exporter for [TP-Link's Tapo smart plugs](https://www.tp-link.com/uk/home-networking/smart-plug/) (primarily for the energy monitoring).

I use this for my [Tapo P110](https://www.tapo.com/uk/product/smart-plug/tapo-p110/)'s, [Tapo P110M](https://www.tp-link.com/uk/home-networking/smart-plug/tapo-p110m/)'s & [Tapo P304M](https://www.tp-link.com/uk/home-networking/smart-plug/tapo-p304m/)'s.

## 📥 Usage

Download the [latest release](https://github.com/viral32111/tapo-smart-plugs/releases/latest) for your platform. There are prebuilt binaries for Linux and Windows, covering both 32-bit and 64-bit architectures of x86 and ARM. There are additional Linux builds to accommodate glibc and musl systems.

### 🐳 Docker

There are [Docker images](https://github.com/users/viral32111/packages/container/package/tapo-smart-plugs) available for Linux, with variants for glibc (Ubuntu) and musl (Alpine Linux):

 * `ghcr.io/viral32111/tapo-smart-plugs:1.0-musl`
 * `ghcr.io/viral32111/tapo-smart-plugs:1.0-glibc`

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

Create a TOML configuration file similar to the example below.

```toml
# The HTTP server for exposing metrics
[metrics]
listen = "127.0.0.1:5000" # $TAPO_METRICS_LISTEN
path = "/metrics" # $TAPO_METRICS_PATH
interval = "10s"

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

Core configuration (not registering of devices) is possible via environment variables for easy deployment with Docker. These take priority over the TOML configuration file. See [`.env.example`](/.env.example).

| Name | Type | Description | Default |
| --- | --- | --- | --- |
| `TAPO_METRICS_LISTEN` | String | IP address & TCP port to listen for HTTP requests on. | `127.0.0.1:5000` |
| `TAPO_METRICS_PATH` | String | HTTP path to expose metrics at. | `/metrics` |
| `TAPO_METRICS_INTERVAL` | String | Interval to refresh metrics. | `10s` |
| `TAPO_ACCOUNT_EMAIL` | String | Email address of the Tapo account. | |
| `TAPO_ACCOUNT_PASSWORD` | String | Password of the Tapo account. | |

## 🧰 Development

1. Copy the [`.env.example`](/.env.example) file to `.env` and populate accordingly.
2. Run the application with `go run ./cmd/tapo/main.go`.

## 🛠️ Building

Run the relevant build command below, assuming you're on a Linux 64-bit (glibc) system.

### Linux 64-bit (glibc)

```bash
GOOS=linux GOARCH=amd64 \
	go build -o tapo-smart-plugs-v0.0.0-linux-amd64-glibc ./cmd/tapo/main.go
```

### Linux 64-bit (musl)

```bash
CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
	go build -o tapo-smart-plugs-v0.0.0-linux-amd64-musl ./cmd/tapo/main.go
```

### Windows 64-bit

```bash
GOOS=windows GOARCH=amd64 \
	go build -o tapo-smart-plugs-v0.0.0-windows-amd64.exe ./cmd/tapo/main.go
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
