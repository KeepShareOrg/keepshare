# Changelog

## Unreleased

- feat: forward PikPak master-account-bound noreply emails (from
  `noreply@accounts.mypikpak.com`) to the matching keepshare user via a rotating
  SMTP pool configured under `forward_mails`. Disabled by default. See
  `docs/forward.md`.
