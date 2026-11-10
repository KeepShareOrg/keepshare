# Changelog

## Unreleased

- Reject ed2k file links without a valid `|/` terminator; keep later batch links
  separate when an earlier ed2k link is incomplete.
- feat: forward PikPak master-account-bound noreply emails (from
  `noreply@accounts.mypikpak.com`) to the matching keepshare user via a rotating
  SMTP pool configured under `forward_mails`. Disabled by default. See
  `docs/forward.md`.
- Include the HTTP `Referer` header as `refer` in `visit_link` request logs.
- XLCO-13652: Add per-link landing-page Page Views and Visitor IPs to Shared Links,
  column selection, clipboard copy, and Excel export. Rename Visitor to Host
  Visitors without changing hosting-provider statistics.
- Count successful landing document loads, not direct hosting-provider redirects
  or status polling. Persist exact IP deduplication with link-scoped fingerprints
  and atomic counters; remove fingerprints when a sharing record is deleted.
- Add the existing-install migration `migrations/20261106_shared_link_visits.sql`.
  Apply it before deploying the backend; historical counters start at zero.
  Require instant DDL (MySQL 8.0.29+) and a five-second metadata-lock timeout to
  avoid an unexpected rebuild of large sharing tables.
- Keep Gin's default client IP handling without extra proxy configuration.
  Forwarded headers are trusted, so Visitor IPs is a reference metric rather than
  a security or billing signal. Preserve uncached landing documents and existing
  page responses when statistics writes fail.
