# KeepShare - File hosting and sharing automation.

<img width="240" src="docs/logo.png">

[![Release](https://img.shields.io/github/release/KeepShareOrg/keepshare.svg?style=flat-square)](https://github.com/KeepShareOrg/keepshare/releases)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)](CODE_OF_CONDUCT.md)

## Table of Contents

- [Introduction](#introduction)
- [Getting started with our public server](#getting-started-with-our-public-server)
- [Getting started with your private server (Self-Host)](#getting-started-with-your-private-server-self-host)
  - [Run with Docker](#run-with-docker)
  - [Compile from Source (for development)](#compile-from-source-for-development)
      - [Prerequisites](#prerequisites)
      - [Build and Run](#build-and-run)
      - [Configuration](#configuration)
  - [Usage](#usage)
- [Documentation](#documentation)
- [Supported file hosting providers](#supported-file-hosting-providers)
- [Roadmap](#roadmap)
- [Code of Conduct](#code-of-conduct)
- [Contributing](#contributing)
- [Donation](#donation)
- [License](#license)

## Introduction

KeepShare is a tool for automated and batch file hosting and sharing. With KeepShare, you can easily create a large number of file shares through uploads or links such as DDL/Megent/Ed2K, and automatically keep the shares valid.

Why KeepShare?
- Quickly eliminate the "original sin" of Magnet links
- Anonymous file sharing publisher
- Unlimited file hosting
- Automatically repair banned sharing
- Help you make money, not us

Goals:
- Open and Transparent
- Automation
- Easy Integration
- Controlled by You
- Easy to Self-Host

Please go to [keepshare.org/docs/intro/](https://keepshare.org/docs/intro/) for details.

## Getting started with our public server

`RECOMMENDED`

1. Go to [keepshare.org/console](https://keepshare.org/console) to create an account and log in.
2. Combine your download links according to the `Auto-Share Link Template` to get `Keep Sharing Links`.
3. Post your `Keep Sharing Links`...

Please log in to the console to learn more features and usage, it's very simple.

## Getting started with your private server (Self-Host)

You can choose one of [Run with Docker](#run-with-docker) or [Compile from Source](#compile-from-source-for-development) to build a self-host service.

### Run with Docker

```
docker run \
  -itd \
  -e KS_ROOT_DOMAIN=keepshare.org \
  -e KS_DB_MYSQL='user:password@(127.0.0.1:3306)/keepshare?parseTime=True&loc=Local' \
  -e KS_DB_REDIS='redis://localhost:6379?max_retries=2' \
  keepshare/keepshare@latest
```

### Compile from Source (for development)

#### Prerequisites

- mysql 8.0+
- redis 5.0+
- golang 1.19+
- nodejs 18+
- pnpm 8+

#### Build and Run

``` bash
make build-fe
make build

# create mysql database.
mysql -uroot -padmin -h127.0.0.1 -P3306 -e 'CREATE DATABASE keepshare'

# create mysql tables.
./keepshare tables create

# show configurations
./keepshare config

# start server
./keepshare start
```

#### Configuration

Run `./keepshare config` to view details.

Client IPs use Gin's default behavior: `X-Forwarded-For`, then `X-Real-IP`,
falling back to the connection's remote address. No extra proxy configuration is
required. These headers are trusted without source checks, so Visitor IPs is a
reference metric that can be inflated with forged headers, not a security or
billing signal.

### Usage

Same as [Getting started with our public server](#getting-started-with-our-public-server), except replace the console page address with the one you hosted and configured.

### Shared Links landing-page statistics

- **Page Views** counts successful GET document loads of `/console/shared/status`
  and `/console/shared/wsl-status` for the sharing record, including refreshes.
  Direct redirects to the hosting provider, status API polls, and static assets do
  not count. These are page requests, not verified human visitors.
- **Visitor IPs** counts distinct normalized IP addresses over the record's
  lifetime. Equivalent IPv4/mapped IPv6 addresses share one visitor. Invalid IPs
  do not increase this metric, but the page view still counts. Only link-scoped
  SHA-256 fingerprints are retained for deduplication; this table stores no raw IPs.
- **Host Visitors** is the existing hosting-provider `visitor` statistic, unchanged.
- Both new columns are visible by default and available in column selection,
  clipboard copy, and Excel export. Historical records start at zero when upgraded.
- Landing documents use `Cache-Control: no-store`; do not override this policy in
  a CDN or reverse proxy. A statistics write failure is logged without breaking
  the page response. Deleting a sharing record also deletes its IP fingerprints.

For an existing installation, apply `migrations/20261106_shared_link_visits.sql`
once to the KeepShare MySQL database **before** deploying the updated backend,
then deploy the frontend. For example:

```sh
mysql --host=YOUR_DB_HOST --user=YOUR_DB_USER --password YOUR_DB_NAME < migrations/20261106_shared_link_visits.sql
```

`keepshare tables create` includes the new schema for fresh installations, but
does not alter existing tables. Back up the database before migration. Roll back
application binaries without dropping the new columns or fingerprint table.
The migration requires MySQL 8.0.29+ for instant column insertion and limits
metadata-lock waits to five seconds. Unsupported instant DDL or lock contention
fails instead of falling back to a rebuild of the sharing table.

Counter and migration regression tests use an isolated temporary database and
require MySQL create/drop-database privileges:

```sh
KEEPSHARE_TEST_MYSQL_DSN='user:password@tcp(127.0.0.1:3306)/?parseTime=true' \
  go test ./server -run 'Test(LinkVisitorIPHash|RecordLinkVisit.*|LinkVisitsExistingInstallMigrationMySQL|LandingPageDocumentVisitsMySQL|ProxiedLandingPageVisitsMySQL)$'
```

## Documentation

TODO - [keepshare.org/docs](https://keepshare.org/docs) [WIP]

## Supported file hosting providers

- [PikPak](https://mypikpak.com/)
- TODO: [RapidGator](https://rapidgator.net/)
- TODO: [Mega](https://mega.io/)

We hope to support as many file hosting providers as possible, pull requests are welcome!

## Roadmap

Please see [ROADMAP.md](ROADMAP.md) for details, it will be updated as the project proceeds.

## Code of Conduct

Help us keep open and inclusive. Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## Contributing

KeepShare is the work of many contributors. We appreciate your help!

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on submitting patches and the contribution workflow.

Thanks goes to the wonderful people listed in [AUTHORS.md](AUTHORS.md).

## Donation

If KeepShare helps you a lot, you can support us by donate premium redemption codes of file hosting providers at [keepshare.org/donate](https://keepshare.org/donate) [WIP].

Yes, we do not need you to donate money, but would appreciate premium redemption codes issued by [supported file hosting providers](#supported-file-hosting-providers) as a donation. These redemption codes can help more KeepShare users create more keep sharing links, making the KeepShare system more healthy and vital.

## License

The code in this repository is released under the MIT License.
