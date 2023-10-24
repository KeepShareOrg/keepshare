CREATE TABLE IF NOT EXISTS `keepshare_blacklist`
(
    `user_id`            varchar(16) NOT NULL,
    `original_link_hash` char(40)    NOT NULL,
    `original_link`      text        NOT NULL,
    `created_at`         datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY `user_id.original_link_hash` (`user_id`, `original_link_hash`),
    KEY `user_id.created_at` (`user_id`, `created_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_bin;
