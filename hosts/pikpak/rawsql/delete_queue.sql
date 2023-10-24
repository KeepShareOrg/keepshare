CREATE TABLE IF NOT EXISTS `pikpak_delete_queue`
(
    `worker_user_id`     varchar(20) NOT NULL,
    `original_link_hash` char(40)    NOT NULL,
    `status`             varchar(32) NOT NULL,
    `created_at`         datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `next_trigger`       datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `ext`                text,
    PRIMARY KEY (`worker_user_id`, `original_link_hash`),
    KEY (`status`, `next_trigger`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_bin;
