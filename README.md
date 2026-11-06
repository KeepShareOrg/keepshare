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
  - [Global magnet blacklist](#global-magnet-blacklist)
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
### Global magnet blacklist

Operators with access to the service configuration and database can block a magnet across all users and channels. This is separate from each user's Link Blacklist. The commands use the existing `KS_DB_MYSQL` / `KS_DB_REDIS` environment variables or `./conf/config.yaml` / `/etc/keepshare/config.yaml`; they do not start the HTTP server or background workers.

Deploy in this order:

1. Build the frontend (`cd web && pnpm install --frozen-lockfile && pnpm run build`), then build the backend from the repository root (`go build -mod=mod -o keepshare .`).
2. Before starting the new backend, create the additional table using its configured database:
   ```sh
   ./keepshare tables create keepshare_magnet_blacklist
   ```
   For an **existing** `pikpak_file` table, apply the key-width migration to the same database before deploying the backend:
   ```sh
   mysql --host=127.0.0.1 --user=keepshare --password keepshare < migrations/20261106_pikpak_file_unique_hash.sql
   ```
   Replace the connection options/database with the configured values. The migration preserves existing rows and indexes and widens `unique_hash` from 50 to 57 characters (16-character user ID + `:` + 40-character BTIH). Re-running it is safe; schedule the `ALTER TABLE` for an appropriate maintenance window because it can wait on table locks. Fresh installations use the new width automatically. `tables create` skips existing tables and does **not** apply this migration. Do not use `--drop-empty` for this deployment.
3. Deploy the new backend and frontend to **all** instances before using the blacklist commands. Older instances do not enforce global blocks.
4. Check a blocked resource through its auto-share URL, old status-page URL, and `/api/shared_link?id=...`; it must show resource information without returning a PikPak share URL. Also check an unblocked base32 magnet and a status page whose Whatslink lookup takes longer than a polling interval.

```sh
./keepshare magnet-blacklist add 'magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567'
./keepshare magnet-blacklist list
./keepshare magnet-blacklist remove 'magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567'
```

Quote the entire URI, including any `&dn=` or tracker parameters. BTIH hexadecimal/base32 encoding, letter case, display names, and trackers do not change the blocked resource. Invalid or conflicting BTIH topics are rejected before connecting to the database.

All commands emit JSON on stdout and return a nonzero exit status on failure. `add` first persists the global block, then revokes matching native PikPak shares across accounts. It **does not delete cloud files**. Successful revocations clear local share URLs/status caches while preserving personal blocks and resource display records. The read-only resource page queries Whatslink, not PikPak.

An `add` result can contain `local_blocked: true` and `error.code: "revocation_incomplete"`: local access is already blocked, but some native shares or cleanup operations failed. Inspect `remote.failures` (`worker_user_id`, `share_id`, and `stage`), restore the worker credentials or database/Redis availability, and run the same `add` again. It retries remaining mappings even if the magnet is already blocked. A `mapping` failure requires recovering the account/share mapping; the tool does not guess a worker or claim an untracked share was revoked. Preserve the failure output for that recovery.

Revocation is not atomic with the remote service. An in-flight share creation is checked again on completion and compensated by revoking the new share; failed compensation remains discoverable for a repeated `add`. Already distributed native URLs remain usable until their remote revocation succeeds. Successful revocation is irreversible: `remove` only lifts the global block; it does not restore old share URLs, remove personal/channel restrictions, or change cloud files. Database errors fail closed rather than creating or disclosing a share.

Native-share publication and revocation cleanup lock the same share mapping. A callback arriving after successful revocation cannot restore the old URL even if the global block has since been removed; a newly created native share can still complete after unblocking. Status polling and Whatslink metadata loading have separate lifetimes, so a slow metadata response is retained while polling continues.

The MySQL migration/publication regressions use `KS_TEST_MYSQL_DSN` and create/drop uniquely named temporary databases; point this variable only at a disposable MySQL instance with `CREATE DATABASE` permission:
```sh
KS_TEST_MYSQL_DSN='root@tcp(127.0.0.1:3306)/keepshare?parseTime=True&loc=Local' \
  go test -vet=off ./server -run 'Test(MagnetTaskKeyMigration|RevokedNativeShare|NativePublication)'
```
Without the variable, these database-dependent tests are skipped.

Do not roll back to a backend that does not enforce the blacklist while any global blocks remain active. Deploy the table first: a missing table intentionally prevents magnet access instead of silently allowing it.

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
