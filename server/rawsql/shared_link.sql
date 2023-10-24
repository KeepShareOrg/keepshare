CREATE TABLE IF NOT EXISTS `keepshare_shared_link`
(
    `auto_id`               bigint       NOT NULL AUTO_INCREMENT,
    `user_id`               varchar(16)  NOT NULL,
    `state`                 varchar(20)  NOT NULL DEFAULT '',
    `host`                  varchar(20)  NOT NULL DEFAULT '',
    `created_by`            varchar(20)  NOT NULL DEFAULT '',
    `created_at`            datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`            datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `size`                  bigint       NOT NULL DEFAULT 0,
    `visitor`               int          NOT NULL DEFAULT 0,
    `stored`                int          NOT NULL DEFAULT 0,
    `last_visited_at`       datetime     NOT NULL DEFAULT '2000-01-01 00:00:00',
    `last_stored_at`        datetime     NOT NULL DEFAULT '2000-01-01 00:00:00',
    `revenue`               bigint       NOT NULL DEFAULT 0,
    `title`                 varchar(256) NOT NULL DEFAULT '',
    `original_link_hash`    char(40)     NOT NULL,
    `host_shared_link_hash` char(40)     NOT NULL,
    `original_link`         text         NOT NULL,
    `host_shared_link`      text         NOT NULL,
    PRIMARY KEY (`auto_id`),
    KEY `user_id.created_at` (`user_id`, `created_at`),
    KEY `host_shared_link_hash.user_id` (`host_shared_link_hash`, `user_id`),
    UNIQUE KEY `original_link_hash.user_id` (`original_link_hash`, `user_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_bin;
