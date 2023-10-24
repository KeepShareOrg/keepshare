CREATE TABLE IF NOT EXISTS `pikpak_worker_account`
(
    `user_id`          varchar(20) NOT NULL,
    `master_user_id`   varchar(20) NOT NULL DEFAULT '',
    `email`            varchar(64) NOT NULL,
    `password`         varchar(64) NOT NULL,
    `used_size`        bigint      NOT NULL,
    `limit_size`       bigint      NOT NULL,
    premium_expiration datetime    NOT NULL DEFAULT '2000-01-01 00:00:00',
    `created_at`       datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`       datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`user_id`),
    KEY `master_user_id` (`master_user_id`),
    KEY `premium_expiration` (premium_expiration)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_bin;
