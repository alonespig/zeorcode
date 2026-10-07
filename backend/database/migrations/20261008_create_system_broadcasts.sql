CREATE TABLE IF NOT EXISTS `system_broadcasts` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `actor_id` bigint NOT NULL,
  `title` varchar(255) NOT NULL,
  `content` varchar(2000) NOT NULL,
  `link` varchar(255) NOT NULL DEFAULT '',
  `recipient_count` int NOT NULL DEFAULT '0',
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_system_broadcasts_actor_id` (`actor_id`),
  KEY `idx_system_broadcasts_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
