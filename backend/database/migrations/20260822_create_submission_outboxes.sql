-- Submission 与判题队列之间的事务 Outbox。
-- 上线顺序：先执行本迁移，再部署写入 Outbox 的应用版本。

CREATE TABLE IF NOT EXISTS `submission_outboxes` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `submission_id` bigint NOT NULL,
  `submission_version` bigint NOT NULL,
  `status` tinyint unsigned NOT NULL DEFAULT '0',
  `attempts` bigint NOT NULL DEFAULT '0',
  `available_at` datetime(3) NOT NULL,
  `locked_by` varchar(128) NOT NULL DEFAULT '',
  `locked_until` datetime(3) DEFAULT NULL,
  `published_at` datetime(3) DEFAULT NULL,
  `last_error` varchar(1024) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_submission_outbox_version` (`submission_id`,`submission_version`),
  KEY `idx_submission_outbox_dispatch` (`status`,`available_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
