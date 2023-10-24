CREATE TABLE IF NOT EXISTS `keepshare_user`
(
    `id`             varchar(16) NOT NULL,
    `name`           varchar(64) NOT NULL,
    `email`          varchar(64) NOT NULL DEFAULT '',
    `password_hash`  char(64)    NOT NULL,
    `channel`        varchar(32) NOT NULL,
    `email_verified` int         NOT NULL DEFAULT 0, # 0: not verified, 1: verified
    `created_at`     datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `email` (`email`),
    UNIQUE KEY `channel` (`channel`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_bin;
