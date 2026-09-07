-- ZeorCode Agent durable conversations and auditable write actions.
-- Redis is only used for short-lived run locks; conversation state remains in MySQL.

CREATE TABLE `ai_conversations` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `public_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `title` varchar(120) NOT NULL DEFAULT '新对话',
  `summary` text,
  `state_json` longtext,
  `status` int NOT NULL DEFAULT '0',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ai_conversations_public_id` (`public_id`),
  KEY `idx_ai_conversation_user_updated` (`user_id`,`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `ai_messages` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `public_id` bigint NOT NULL,
  `conversation_id` bigint NOT NULL,
  `run_id` bigint NOT NULL DEFAULT '0',
  `role` varchar(16) NOT NULL,
  `kind` varchar(24) NOT NULL DEFAULT 'text',
  `content` longtext,
  `blocks_json` longtext,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ai_messages_public_id` (`public_id`),
  KEY `idx_ai_message_conversation_created` (`conversation_id`,`created_at`),
  KEY `idx_ai_messages_run_id` (`run_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `ai_runs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `public_id` bigint NOT NULL,
  `conversation_id` bigint NOT NULL,
  `status` varchar(16) NOT NULL,
  `model` varchar(100) NOT NULL DEFAULT '',
  `error_code` varchar(64) NOT NULL DEFAULT '',
  `started_at` datetime(3) DEFAULT NULL,
  `finished_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ai_runs_public_id` (`public_id`),
  KEY `idx_ai_runs_conversation_id` (`conversation_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `ai_actions` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `public_id` bigint NOT NULL,
  `conversation_id` bigint NOT NULL,
  `run_id` bigint NOT NULL DEFAULT '0',
  `action_type` varchar(40) NOT NULL,
  `artifact_json` longtext NOT NULL,
  `artifact_version` int NOT NULL,
  `payload_hash` varchar(64) NOT NULL,
  `status` varchar(16) NOT NULL,
  `result_public_id` bigint NOT NULL DEFAULT '0',
  `approved_by` bigint NOT NULL DEFAULT '0',
  `approved_at` datetime(3) DEFAULT NULL,
  `error_code` varchar(64) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ai_actions_public_id` (`public_id`),
  KEY `idx_ai_actions_conversation_id` (`conversation_id`),
  KEY `idx_ai_actions_run_id` (`run_id`),
  KEY `idx_ai_actions_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- The source action is the idempotency key for a confirmed Agent write.
ALTER TABLE `homeworks`
  ADD COLUMN `source_action_id` bigint DEFAULT NULL AFTER `created_by`,
  ADD UNIQUE KEY `idx_homeworks_source_action_id` (`source_action_id`);
