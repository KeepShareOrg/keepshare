-- Apply once to the existing KeepShare database before starting the updated server.
-- Existing sharing records start at zero, and host visitor statistics are unchanged.
-- Require instant DDL rather than falling back to a table rebuild.
SET SESSION lock_wait_timeout = 5;
SET SESSION innodb_lock_wait_timeout = 5;

ALTER TABLE `keepshare_shared_link`
    ADD COLUMN `page_views` bigint NOT NULL DEFAULT 0 AFTER `visitor`,
    ADD COLUMN `visitor_ips` bigint NOT NULL DEFAULT 0 AFTER `page_views`,
    ALGORITHM = INSTANT;

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
