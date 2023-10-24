CREATE TABLE IF NOT EXISTS `pikpak_file`
(
    `auto_id`            bigint       NOT NULL AUTO_INCREMENT,
    `master_user_id`     varchar(20)  NOT NULL,
    `worker_user_id`     varchar(20)  NOT NULL,
    `file_id`            varchar(32)  NOT NULL,
    `task_id`            varchar(32)  NOT NULL,
    `status`             varchar(32)  NOT NULL,
    `is_dir`             bool         NOT NULL DEFAULT false,
    `size`               bigint       NOT NULL DEFAULT 0,
    `name`               varchar(256) NOT NULL DEFAULT '',
    `created_at`         datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`         datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `original_link_hash` char(40)     NOT NULL,
    PRIMARY KEY (`auto_id`),
    UNIQUE KEY (`master_user_id`, `original_link_hash`),
    UNIQUE KEY (`worker_user_id`, `original_link_hash`),
    KEY (`task_id`),
    KEY (`status`, `updated_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_bin;
