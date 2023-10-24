CREATE TABLE IF NOT EXISTS `pikpak_shared_link`
(
    `share_id`       varchar(32) NOT NULL,
    `file_id`        varchar(32) NOT NULL,
    `master_user_id` varchar(20) NOT NULL,
    `worker_user_id` varchar(20) NOT NULL,
    `created_at`     datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`share_id`),
    KEY (`file_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_bin;
