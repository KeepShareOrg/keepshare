CREATE TABLE IF NOT EXISTS `pikpak_token`
(
    `user_id`       varchar(20)   NOT NULL,
    `access_token`  varchar(2048) NOT NULL,
    `refresh_token` varchar(2048) NOT NULL,
    `expiration`    datetime      NOT NULL,
    `created_at`    datetime      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`user_id`),
    KEY `expiration` (`expiration`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_bin;
