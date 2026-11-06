CREATE TABLE IF NOT EXISTS `keepshare_magnet_blacklist`
(
    `info_hash`     char(40) NOT NULL,
    `original_link` text     NOT NULL,
    `created_at`    datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`info_hash`),
    KEY `created_at.info_hash` (`created_at`, `info_hash`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_bin;
