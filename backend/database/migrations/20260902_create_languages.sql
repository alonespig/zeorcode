-- 编程语言配置：数据库控制前台是否可选，具体编译/运行命令由 pkg/judge 安全预设提供。
CREATE TABLE IF NOT EXISTS `languages` (
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

-- 固定已有编号，兼容浏览器保存的语言选择以及现有提交接口。
INSERT IGNORE INTO `languages` (`id`, `name`, `status`, `sort`, `created_at`, `updated_at`) VALUES
  (1, 'c++', 1, 1, NOW(3), NOW(3)),
  (2, 'java', 1, 2, NOW(3), NOW(3)),
  (3, 'python', 1, 3, NOW(3), NOW(3));
