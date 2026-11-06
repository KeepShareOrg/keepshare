CREATE TABLE IF NOT EXISTS `keepshare_shared_link_visitor_ip`
(
    `shared_link_id` bigint     NOT NULL,
    `ip_hash`        binary(32) NOT NULL,
    PRIMARY KEY (`shared_link_id`, `ip_hash`),
    CONSTRAINT `fk_shared_link_visitor_ip_link`
        FOREIGN KEY (`shared_link_id`) REFERENCES `keepshare_shared_link` (`auto_id`) ON DELETE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_bin;
