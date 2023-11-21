# `baton-sentinel-one` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-sentinel-one.svg)](https://pkg.go.dev/github.com/conductorone/baton-sentinel-one) ![main ci](https://github.com/conductorone/baton-sentinel-one/actions/workflows/main.yaml/badge.svg)

`baton-sentinel-one` is a connector for SentinelOne built using the [Baton SDK](https://github.com/conductorone/baton-sdk). It communicates with the SentinelOne API to sync data about users, service users, sites, roles and accounts.
Check out [Baton](https://github.com/conductorone/baton) to learn more about the project in general.

# Getting Started

## Prerequisites

1. Access to SentinelOne management console.
2. API key and the management console url.

- management console url is used to call API endpoints to fetch data from SentinelOne. E.g https://some-url.sentinelone.net
- to create an API key go to your management console, click on your name in top right corner then select My User - Actions - Api Token Operations - Generate API Token

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-sentinel-one

BATON_API_TOKEN=sentinelOneApiToken BATON_MANAGEMENT_CONSOLE_URL=https://your-management-url.net baton-sentinel-one
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out -e BATON_API_TOKEN=sentinelOneApiToken BATON_MANAGEMENT_CONSOLE_URL=https://your-management-url.net ghcr.io/conductorone/baton-sentinel-one:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-sentinel-one/cmd/baton-sentinel-one@main

BATON_API_TOKEN=sentinelOneApiToken BATON_MANAGEMENT_CONSOLE_URL=https://your-management-url.net baton-sentinel-one
baton resources
```

# Data Model

`baton-sentinel-one` will pull down information about the following SentinelOne resources:

- Accounts
- Users
- Service users
- Sites
- Roles

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually building spreadsheets. We welcome contributions, and ideas, no matter how small -- our goal is to make identity and permissions sprawl less painful for everyone. If you have questions, problems, or ideas: Please open a Github Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-sentinel-one` Command Line Usage

```
baton-sentinel-one

Usage:
  baton-sentinel-one [flags]
  baton-sentinel-one [command]

Available Commands:
  completion         Generate the autocompletion script for the specified shell
  help               Help about any command

Flags:
      --api-token string                API token for your management console used to authenticate with SentinelOne API. ($BATON_API_TOKEN)
      --client-id string                The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string            The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
  -f, --file string                     The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                            help for baton-sentinel-one
      --log-format string               The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string                The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
      --management-console-url string   Your management console url. ($BATON_MANAGEMENT_CONSOLE_URL)
  -v, --version                         version for baton-sentinel-one

Use "baton-sentinel-one [command] --help" for more information about a command.
```
