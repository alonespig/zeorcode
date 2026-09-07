-- 题单（ProblemSet）：管理员挑一组题配上 Markdown 说明，用户按单刷题并看到整体进度。
-- 本项目不做 AutoMigrate，表结构由 schema.sql + 本目录下的迁移脚本维护。
-- 上线顺序：先执行本迁移，再部署带题单接口的应用版本。

-- 题单主表。
-- published 与 visibility 是两个正交的轴，不要合并：
--   published  管「编好之前不给看」（0 草稿 / 1 已发布）
--   visibility 管「给谁看」（0 公开 / 1 需邀请码）
CREATE TABLE IF NOT EXISTS `problem_sets` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `title` varchar(100) NOT NULL,
  `description` text,
  `published` tinyint NOT NULL DEFAULT '0',
  `visibility` tinyint NOT NULL DEFAULT '0',
  `invite_code` varchar(32) NOT NULL DEFAULT '',
  `created_by` bigint DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_problem_sets_created_by` (`created_by`),
  KEY `idx_problem_sets_listing` (`published`,`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 题单题目关联，sort 决定展示顺序（从 0 递增）。
-- 唯一键防同一题在同一题单里重复出现。
CREATE TABLE IF NOT EXISTS `problem_set_problems` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `problem_set_id` bigint DEFAULT NULL,
  `problem_id` bigint DEFAULT NULL,
  `sort` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_set_problem` (`problem_set_id`,`problem_id`),
  KEY `idx_problem_set_problems_set_id` (`problem_set_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 题单标签关联，复用题目那套 tags 表。
CREATE TABLE IF NOT EXISTS `problem_set_tags` (
  `problem_set_id` bigint DEFAULT NULL,
  `tag_id` bigint DEFAULT NULL,
  KEY `idx_set_tag` (`problem_set_id`,`tag_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 邀请码题单的解锁记录：输对一次后长期有效，不必每次重输。
CREATE TABLE IF NOT EXISTS `problem_set_unlocks` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint DEFAULT NULL,
  `problem_set_id` bigint DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_set_unlock` (`user_id`,`problem_set_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
