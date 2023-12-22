CREATE TABLE IF NOT EXISTS `pikpak_redeem_code`
(
    `auto_id`      bigint   NOT NULL AUTO_INCREMENT,
    `code`         char(40) NOT NULL DEFAULT '',
    `status`       char(10) NOT NULL DEFAULT 'NOT_USED' COMMENT 'NOT_USED, USED, INVALID',
    `used_user_id` char(20)          DEFAULT '',
    `created_at`   datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`auto_id`),
    KEY (`code`),
    KEY (`used_user_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_bin;
