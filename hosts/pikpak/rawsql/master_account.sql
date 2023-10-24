CREATE TABLE IF NOT EXISTS `pikpak_master_account`
(
    `user_id`           varchar(20) NOT NULL,
    `keepshare_user_id` varchar(16) NOT NULL DEFAULT '',
    `email`             varchar(64) NOT NULL,
    `password`          varchar(64) NOT NULL,
    `created_at`        datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`        datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`user_id`),
    KEY `keepshare_user_id` (`keepshare_user_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_bin;
