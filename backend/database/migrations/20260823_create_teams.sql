-- 团队（Team）与团队作业（Homework）。
-- 本项目不做 AutoMigrate，表结构由 schema.sql + 本目录下的迁移脚本维护。
-- 上线顺序：先执行本迁移，再部署带团队接口的应用版本。
--
-- 本轮只做「团队 + 作业核心」，团队文件（team_files）留到下一轮，故这里不建。

-- 团队主表。visibility：0 公开（任何人可直接加入）/ 1 非公开（需邀请码）。
CREATE TABLE IF NOT EXISTS `teams` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(60) NOT NULL,
  `description` text,
  `visibility` tinyint NOT NULL DEFAULT '0',
  `invite_code` varchar(32) NOT NULL DEFAULT '',
  `owner_id` bigint DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_teams_owner_id` (`owner_id`),
  KEY `idx_teams_listing` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 团队成员。role：0 成员 / 1 团队管理员 / 2 所有者。
-- remark 是历史备注名字段；20260826 迁移会将其回填到 users.real_name 后删除。
CREATE TABLE IF NOT EXISTS `team_members` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `team_id` bigint DEFAULT NULL,
  `user_id` bigint DEFAULT NULL,
  `role` tinyint NOT NULL DEFAULT '0',
  `remark` varchar(30) NOT NULL DEFAULT '',
  `joined_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_team_member` (`team_id`,`user_id`),
  KEY `idx_team_members_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 团队作业。start_time/end_time 构成计分时间窗：
-- 截止后仍可提交，但窗外的提交不计入排行榜。
-- created_by 决定谁能改：只有本人能改（团队管理员也不行），删除额外给团队所有者兜底。
CREATE TABLE IF NOT EXISTS `homeworks` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `team_id` bigint DEFAULT NULL,
  `title` varchar(100) NOT NULL,
  `description` text,
  `start_time` datetime(3) DEFAULT NULL,
  `end_time` datetime(3) DEFAULT NULL,
  `created_by` bigint DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_homework_team_start` (`team_id`,`start_time`),
  KEY `idx_homeworks_created_by` (`created_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 作业题目关联，sort 决定顺序。唯一键防同题在同一作业里重复。
-- 不存分值：作业每题满分统一 100（model.HomeworkProblemFullScore）。
CREATE TABLE IF NOT EXISTS `homework_problems` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `homework_id` bigint DEFAULT NULL,
  `problem_id` bigint DEFAULT NULL,
  `sort` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_homework_problem` (`homework_id`,`problem_id`),
  KEY `idx_homework_problems_homework_id` (`homework_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 提交表加 homework_id，与 contest_id 对称、互斥（0 表示非作业提交）。
-- 作业不是比赛，塞进 contest 会污染 rating/封榜/赛制那一整套逻辑，所以单开一列。
-- 复合索引同时服务两个查询：
--   排行榜  WHERE homework_id=? AND created_at BETWEEN ? AND ?  （用 homework_id 前缀）
--   提交列表 WHERE homework_id=? AND user_id=? ORDER BY created_at DESC
ALTER TABLE `submissions`
  ADD COLUMN `homework_id` bigint NOT NULL DEFAULT '0' AFTER `contest_id`,
  ADD KEY `idx_submission_homework_user` (`homework_id`,`user_id`,`created_at`);
