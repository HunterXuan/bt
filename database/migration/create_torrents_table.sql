CREATE TABLE `torrents` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `created_at` datetime(3) DEFAULT NULL,
    `updated_at` datetime(3) DEFAULT NULL,
    `info_hash` varchar(40) COLLATE utf8mb4_unicode_ci NOT NULL,
    `seeder_count` bigint(20) unsigned NOT NULL,
    `leecher_count` bigint(20) unsigned NOT NULL,
    `snatcher_count` bigint(20) unsigned NOT NULL,
    `last_active_at` datetime(3) DEFAULT NULL,
    `meta_info` mediumtext COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
    PRIMARY KEY (`id`),
    UNIQUE KEY `info_hash` (`info_hash`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;