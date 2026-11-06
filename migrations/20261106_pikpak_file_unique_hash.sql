-- Run against the configured KeepShare database before deploying this version.
-- Keep existing keys and indexes; 16 user-ID characters + ':' + 40 hash characters.
ALTER TABLE `pikpak_file`
    MODIFY COLUMN `unique_hash` char(57) NOT NULL COMMENT 'format: keepshare_uid:original_link_hash';
