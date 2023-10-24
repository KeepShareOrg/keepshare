# KeepShare - File hosting and sharing automation.

<img width="240" src="docs/logo.png">

[![Release](https://img.shields.io/github/release/KeepShareOrg/keepshare.svg?style=flat-square)](https://github.com/KeepShareOrg/keepshare/releases)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)](CODE_OF_CONDUCT.md)

## Table of Contents

- [Introduction](#introduction)
- [Getting started with our public server](#getting-started-with-our-public-server)
- [Getting started with your private server (Self-Host)](#getting-started-with-your-private-server-self-host)
    - [Prerequisites](#prerequisites)
    - [Run with Docker](#run-with-docker)
    - [Compilation from Source](#compilation-from-source)
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

### Prerequisites

- mysql 8.0+
- redis 5.0+
- golang 1.19+

### Run with Docker

```
docker run \
  -itd \
  -e KS_ROOT_DOMAIN=keepshare.org \
  -e KS_DB_MYSQL='user:password@(127.0.0.1:3306)/keepshare?parseTime=True&loc=Local' \
  -e KS_DB_REDIS='redis://localhost:6379?max_retries=2' \
  keepshare/keepshare@latest
```

### Compilation from Source

``` bash
# build front pages and server.
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

### Configuration

Run `./keepshare config` to view details.

### Usage

Same as [Getting started with our public server](#getting-started-with-our-public-server), except replace the console page address with the one you hosted and configured.

## Documentation

TODO - [keepshare.org/docs](https://keepshare.org/docs) [WIP]

## Supported file hosting providers

- [PikPak](https://mypikpak.com/)
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
