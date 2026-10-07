CREATE TABLE IF NOT EXISTS `admin_audit_logs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `actor_id` bigint NOT NULL,
  `method` varchar(8) NOT NULL,
  `path` varchar(255) NOT NULL,
  `target` varchar(255) NOT NULL DEFAULT '',
  `client_ip` varchar(64) NOT NULL DEFAULT '',
  `success` tinyint(1) NOT NULL DEFAULT '0',
  `code` int NOT NULL DEFAULT '0',
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_admin_audit_logs_actor_id` (`actor_id`),
  KEY `idx_admin_audit_logs_path` (`path`),
  KEY `idx_admin_audit_logs_success` (`success`),
  KEY `idx_admin_audit_logs_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
