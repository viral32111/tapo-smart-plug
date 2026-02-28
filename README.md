# Tapo Smart Plug

[![CI](https://github.com/viral32111/tapo-smart-plug/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/viral32111/tapo-smart-plug/actions/workflows/ci.yml)
[![CodeQL](https://github.com/viral32111/tapo-smart-plug/actions/workflows/codeql.yml/badge.svg)](https://github.com/viral32111/tapo-smart-plug/actions/workflows/codeql.yml)
![GitHub tag (with filter)](https://img.shields.io/github/v/tag/viral32111/tapo-smart-plug?label=Latest)
![GitHub repository size](https://img.shields.io/github/repo-size/viral32111/tapo-smart-plug?label=Size)
![GitHub release downloads](https://img.shields.io/github/downloads/viral32111/tapo-smart-plug/total?label=Downloads)
![GitHub commit activity](https://img.shields.io/github/commit-activity/m/viral32111/tapo-smart-plug?label=Commits)

This is a Prometheus metrics exporter for [TP-Link's Tapo smart plugs](https://www.tp-link.com/uk/home-networking/smart-plug/) (primarily for the energy monitoring).

I use this for my [Tapo P110](https://www.tapo.com/uk/product/smart-plug/tapo-p110/)'s, [Tapo P110M](https://www.tp-link.com/uk/home-networking/smart-plug/tapo-p110m/)'s & [Tapo P304M](https://www.tp-link.com/uk/home-networking/smart-plug/tapo-p304m/)'s.

## Commands

Start during local development: `go run ./cmd/tapo/main.go --log-level debug --config local.toml`

Build for production use:

* Linux 64-bit (glibc): `GOOS=linux GOARCH=amd64 go build -o tapo-v0.1.0-linux-amd64-glibc ./cmd/tapo/main.go`
* Linux 64-bit (musl): `CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o tapo-v0.1.0-linux-amd64-musl ./cmd/tapo/main.go`

## ⚙️ Configuration

Create a TOML configuration file similar to the example below.

```toml
# The HTTP server for exposing metrics
[metrics]
listen = "127.0.0.1:5000" # or TAPO_METRICS_LISTEN environment variable
path = "/metrics" # or TAPO_METRICS_PATH environment variable
interval = "10s"

# The TP-Link Tapo account credentials
# These are never sent to the Internet! They are only used for local authentication with the smart plugs.
[account]
email = "john@example.com" # or TAPO_ACCOUNT_EMAIL environment variable
password = "P4ssw0rd!" # or TAPO_ACCOUNT_PASSWORD environment variable

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
enabled = false # Defaults to true if not specified, set to false to disable monitoring this plug without removing it from the config
name = "Heater"
ip_address = "192.168.10.20"
```

## ⚖️ License

Copyright (C) 2026 [viral32111](https://viral32111.com).

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see https://www.gnu.org/licenses.
