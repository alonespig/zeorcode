-- ZOJ 完整数据库 schema（由当前 oj_db 结构导出，含 hidden / display_id / uid 等全部字段）。
-- AutoMigrate 关闭时，全新环境用它建库：mysql -uroot -p oj_db < database/schema.sql


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
DROP TABLE IF EXISTS `contest_problems`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `contest_problems` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `contest_id` bigint DEFAULT NULL,
  `problem_id` bigint DEFAULT NULL,
  `label` varchar(10) DEFAULT NULL,
  `color` varchar(26) DEFAULT NULL,
  `score` bigint DEFAULT '100',
  PRIMARY KEY (`id`),
  KEY `idx_contest_problems_contest_id` (`contest_id`),
  KEY `idx_contest_problems_problem_id` (`problem_id`),
  KEY `idx_contest_problem_pair` (`contest_id`,`problem_id`),
  KEY `idx_contest_problem_label` (`contest_id`,`label`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `contest_users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `contest_users` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `contest_id` bigint DEFAULT NULL,
  `user_id` bigint DEFAULT NULL,
  `join_time` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_contest_user` (`contest_id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `contests`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `contests` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `public_id` bigint NOT NULL,
  `name` longtext,
  `description` longtext,
  `cover_url` varchar(255) NOT NULL DEFAULT '',
  `type` bigint DEFAULT NULL,
  `start_time` datetime(3) DEFAULT NULL,
  `duration` bigint DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `invite_code` varchar(32) NOT NULL DEFAULT '',
  `end_time` datetime(3) DEFAULT NULL,
  `rated` tinyint(1) DEFAULT '0',
  `settled` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_contests_public_id` (`public_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `ai_conversations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `ai_messages`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `ai_runs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `ai_actions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `homework_problems`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `homework_problems` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `homework_id` bigint DEFAULT NULL,
  `problem_id` bigint DEFAULT NULL,
  `sort` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_homework_problem` (`homework_id`,`problem_id`),
  KEY `idx_homework_problems_homework_id` (`homework_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `homeworks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `homeworks` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `public_id` bigint NOT NULL,
  `team_id` bigint DEFAULT NULL,
  `title` varchar(100) NOT NULL,
  `description` text,
  `start_time` datetime(3) DEFAULT NULL,
  `end_time` datetime(3) DEFAULT NULL,
  `created_by` bigint DEFAULT NULL,
  `source_action_id` bigint DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_homeworks_public_id` (`public_id`),
  KEY `idx_homework_team_start` (`team_id`,`start_time`),
  KEY `idx_homeworks_created_by` (`created_by`),
  UNIQUE KEY `idx_homeworks_source_action_id` (`source_action_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `judge_results`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `judge_results` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `submission_id` bigint DEFAULT NULL,
  `case_id` bigint DEFAULT NULL,
  `status` bigint DEFAULT NULL,
  `time_used` bigint DEFAULT NULL,
  `memory_used` bigint DEFAULT NULL,
  `output_diff` longtext,
  PRIMARY KEY (`id`),
  KEY `idx_judge_results_submission_id` (`submission_id`),
  KEY `idx_judge_results_case_id` (`case_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `languages`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `languages` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(32) NOT NULL,
  `status` tinyint NOT NULL DEFAULT '1',
  `sort` int NOT NULL DEFAULT '0',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_languages_name` (`name`),
  KEY `idx_languages_listing` (`status`,`sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `languages` (`id`, `name`, `status`, `sort`, `created_at`, `updated_at`) VALUES
  (1, 'c++', 1, 1, NOW(3), NOW(3)),
  (2, 'java', 1, 2, NOW(3), NOW(3)),
  (3, 'python', 1, 3, NOW(3), NOW(3));
DROP TABLE IF EXISTS `notifications`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `notifications` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `recipient_id` bigint DEFAULT NULL,
  `is_read` tinyint(1) NOT NULL DEFAULT '0',
  `type` varchar(16) DEFAULT NULL,
  `actor_id` bigint DEFAULT NULL,
  `title` varchar(255) DEFAULT NULL,
  `content` varchar(255) DEFAULT NULL,
  `link` varchar(255) DEFAULT NULL,
  `source_type` varchar(32) DEFAULT NULL,
  `source_id` bigint DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_notif_recipient` (`recipient_id`,`is_read`),
  KEY `idx_notifications_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `post_comment_likes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `post_comment_likes` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `comment_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `created_at` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_comment_like_user` (`comment_id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `post_comments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `post_comments` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `post_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `content` text NOT NULL,
  `created_at` bigint DEFAULT NULL,
  `status` bigint NOT NULL DEFAULT '0',
  `updated_at` bigint DEFAULT NULL,
  `root_id` bigint NOT NULL DEFAULT '0',
  `reply_user_id` bigint NOT NULL DEFAULT '0',
  `like_count` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_post_comments_post_id` (`post_id`),
  KEY `idx_post_comments_user_id` (`user_id`),
  KEY `idx_post_comments_status` (`status`),
  KEY `idx_post_root` (`post_id`,`root_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `post_likes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `post_likes` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `post_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `created_at` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_post_user` (`post_id`,`user_id`),
  UNIQUE KEY `idx_post_like_user` (`post_id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `posts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `posts` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL,
  `category` varchar(16) NOT NULL,
  `problem_id` bigint NOT NULL DEFAULT '0',
  `title` varchar(120) NOT NULL,
  `content` longtext NOT NULL,
  `summary` varchar(255) NOT NULL DEFAULT '',
  `like_count` bigint NOT NULL DEFAULT '0',
  `view_count` bigint NOT NULL DEFAULT '0',
  `comment_count` bigint NOT NULL DEFAULT '0',
  `status` bigint NOT NULL DEFAULT '0',
  `review_status` bigint NOT NULL DEFAULT '0',
  `reject_reason` varchar(255) NOT NULL DEFAULT '',
  `created_at` bigint DEFAULT NULL,
  `updated_at` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_posts_user_id` (`user_id`),
  KEY `idx_posts_category` (`category`),
  KEY `idx_post_category_problem` (`category`,`problem_id`),
  KEY `idx_posts_status` (`status`),
  KEY `idx_posts_review_status` (`review_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `problem_samples`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `problem_samples` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `problem_id` bigint DEFAULT NULL,
  `input` longtext,
  `output` longtext,
  `explain` longtext,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `problem_set_problems`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `problem_set_problems` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `problem_set_id` bigint DEFAULT NULL,
  `problem_id` bigint DEFAULT NULL,
  `sort` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_set_problem` (`problem_set_id`,`problem_id`),
  KEY `idx_problem_set_problems_set_id` (`problem_set_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `problem_set_tags`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `problem_set_tags` (
  `problem_set_id` bigint DEFAULT NULL,
  `tag_id` bigint DEFAULT NULL,
  KEY `idx_set_tag` (`problem_set_id`,`tag_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `problem_set_unlocks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `problem_set_unlocks` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint DEFAULT NULL,
  `problem_set_id` bigint DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_set_unlock` (`user_id`,`problem_set_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `problem_sets`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `problem_sets` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `public_id` bigint NOT NULL,
  `title` varchar(100) NOT NULL,
  `description` text,
  `published` tinyint NOT NULL DEFAULT '0',
  `visibility` tinyint NOT NULL DEFAULT '0',
  `invite_code` varchar(32) NOT NULL DEFAULT '',
  `created_by` bigint DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_problem_sets_public_id` (`public_id`),
  KEY `idx_problem_sets_created_by` (`created_by`),
  KEY `idx_problem_sets_listing` (`published`,`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `problem_tags`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `problem_tags` (
  `problem_id` bigint DEFAULT NULL,
  `tag_id` bigint DEFAULT NULL,
  KEY `idx_problem_tag` (`problem_id`,`tag_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `problems`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `problems` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` longtext,
  `difficulty` bigint DEFAULT NULL,
  `time_limit` bigint DEFAULT NULL,
  `memory_limit` bigint DEFAULT NULL,
  `description` longtext,
  `input_format` longtext,
  `output_format` longtext,
  `hint` longtext,
  `created_at` datetime(3) DEFAULT NULL,
  `oj` varchar(20) NOT NULL DEFAULT '',
  `remote_problem_id` varchar(50) NOT NULL DEFAULT '',
  `hidden` tinyint(1) NOT NULL DEFAULT '0',
  `display_id` varchar(32) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_problems_display_id` (`display_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `rating_changes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `rating_changes` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `contest_id` bigint DEFAULT NULL,
  `user_id` bigint DEFAULT NULL,
  `rank` bigint DEFAULT NULL,
  `old_rating` bigint DEFAULT NULL,
  `new_rating` bigint DEFAULT NULL,
  `delta` bigint DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_rc_contest_user` (`contest_id`,`user_id`),
  KEY `idx_rc_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `remote_accounts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `remote_accounts` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `oj` varchar(20) NOT NULL,
  `auth_type` varchar(20) NOT NULL DEFAULT 'cookie',
  `username` varchar(100) NOT NULL DEFAULT '',
  `secret` text NOT NULL,
  `enabled` tinyint(1) NOT NULL DEFAULT '1',
  `valid` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` bigint DEFAULT NULL,
  `updated_at` bigint DEFAULT NULL,
  `busy` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_remote_accounts_oj` (`oj`),
  KEY `idx_remote_accounts_busy` (`busy`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `submissions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `submissions` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `public_id` bigint NOT NULL,
  `problem_id` bigint DEFAULT NULL,
  `user_id` bigint DEFAULT NULL,
  `contest_id` bigint NOT NULL DEFAULT '0',
  `homework_id` bigint NOT NULL DEFAULT '0',
  `code` longtext NOT NULL,
  `language` longtext NOT NULL,
  `status` bigint NOT NULL,
  `time_used` bigint DEFAULT NULL,
  `memory_used` bigint DEFAULT NULL,
  `compile_output` mediumtext,
  `created_at` datetime(3) DEFAULT NULL,
  `score` bigint NOT NULL DEFAULT '0',
  `version` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_submissions_public_id` (`public_id`),
  KEY `idx_submission_problem_status` (`problem_id`,`status`),
  KEY `idx_submission_contest_problem_status_created` (`contest_id`,`problem_id`,`status`,`created_at`),
  KEY `idx_submission_user_status_created` (`user_id`,`status`,`created_at`),
  KEY `idx_submission_contest_user_created` (`contest_id`,`user_id`,`created_at`),
  KEY `idx_submissions_contest_id` (`contest_id`),
  KEY `idx_submission_contest_created` (`contest_id`,`created_at`),
  KEY `idx_submission_homework_user` (`homework_id`,`user_id`,`created_at`),
  KEY `idx_submission_homework_problem_status` (`homework_id`,`problem_id`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `submission_outboxes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `submission_outboxes` (
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
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `tags`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `tags` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_tags_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `team_members`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `team_members` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `team_id` bigint DEFAULT NULL,
  `user_id` bigint DEFAULT NULL,
  `role` tinyint NOT NULL DEFAULT '0',
  `joined_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_team_member` (`team_id`,`user_id`),
  KEY `idx_team_members_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `teams`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `teams` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `public_id` bigint NOT NULL,
  `name` varchar(60) NOT NULL,
  `cover_url` varchar(255) NOT NULL DEFAULT '',
  `description` text,
  `visibility` tinyint NOT NULL DEFAULT '0',
  `invite_code` varchar(32) NOT NULL DEFAULT '',
  `owner_id` bigint DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_teams_public_id` (`public_id`),
  KEY `idx_teams_owner_id` (`owner_id`),
  KEY `idx_teams_listing` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `user_contest_problem`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_contest_problem` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `contest_id` bigint DEFAULT NULL,
  `user_id` bigint DEFAULT NULL,
  `problem_id` bigint DEFAULT NULL,
  `status` bigint DEFAULT NULL,
  `ac_count` bigint DEFAULT NULL,
  `un_ac_count` bigint DEFAULT NULL,
  `sub_count` bigint DEFAULT NULL,
  `ac_time` datetime(3) DEFAULT NULL,
  `create_time` datetime(3) DEFAULT NULL,
  `score` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ucp_lookup` (`contest_id`,`user_id`,`problem_id`),
  KEY `idx_ucp_contest_problem_status` (`contest_id`,`problem_id`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `user_problems`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_problems` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint DEFAULT NULL,
  `problem_id` bigint DEFAULT NULL,
  `status` bigint DEFAULT '0',
  `ac_count` bigint DEFAULT '0',
  `submit_count` bigint DEFAULT '0',
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_problem` (`user_id`,`problem_id`),
  KEY `idx_user_problem_status` (`user_id`,`status`),
  KEY `idx_problem_user_status` (`problem_id`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(191) NOT NULL,
  `student_no` varchar(32) DEFAULT NULL,
  `real_name` varchar(64) NOT NULL,
  `password` longtext NOT NULL,
  `role` bigint DEFAULT '0',
  `signature` varchar(25) DEFAULT '',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `email` varchar(191) DEFAULT NULL,
  `gender` bigint DEFAULT '1',
  `avatar` varchar(255) DEFAULT '',
  `rating` bigint DEFAULT '0',
  `max_rating` bigint DEFAULT '0',
  `school` varchar(50) DEFAULT '',
  `uid` bigint NOT NULL DEFAULT '0',
  `status` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_users_username` (`username`),
  UNIQUE KEY `idx_users_uid` (`uid`),
  UNIQUE KEY `uni_users_student_no` (`student_no`),
  UNIQUE KEY `uni_users_email` (`email`),
  CONSTRAINT `chk_users_real_name_nonblank` CHECK ((char_length(trim(`real_name`)) > 0))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- ============================================================
-- 初始管理员账号（role=1）。用户名 admin / 密码 admin123（bcrypt cost=10）。
-- ⚠️ 上线前请登录后立刻改密码，或改掉这里的哈希。
-- ============================================================
INSERT INTO `users` (`username`, `real_name`, `password`, `role`, `uid`, `email`, `gender`, `avatar`, `signature`, `school`, `rating`, `max_rating`, `created_at`, `updated_at`)
VALUES ('admin', '系统管理员', '$2a$10$sqWhT5iFfUS7nNHrryfmJuRAZdi9McTVceqCYCrkeaCFbZrv0Rarm', 1, 10000000, 'admin@example.com', 1, '', '', '', 0, 0, NOW(3), NOW(3));

