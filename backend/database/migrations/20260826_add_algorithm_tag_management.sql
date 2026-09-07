-- 算法标签由后台统一维护：名称必填、最长 64 字符且不能重复。
-- 如历史数据存在空名称、超长名称或重名，本迁移会在修改表结构前终止，
-- 请先清理这些数据后重新执行。

CREATE TEMPORARY TABLE `_migration_tag_guard` (
  `invalid_count` bigint NOT NULL,
  CONSTRAINT `chk_tag_guard` CHECK (`invalid_count` = 0)
);

INSERT INTO `_migration_tag_guard` (`invalid_count`)
SELECT COUNT(*)
FROM `tags`
WHERE `name` IS NULL
   OR CHAR_LENGTH(TRIM(`name`)) = 0
   OR CHAR_LENGTH(TRIM(`name`)) > 64;

TRUNCATE TABLE `_migration_tag_guard`;

INSERT INTO `_migration_tag_guard` (`invalid_count`)
SELECT COUNT(*)
FROM (
  SELECT TRIM(`name`)
  FROM `tags`
  GROUP BY TRIM(`name`)
  HAVING COUNT(*) > 1
) AS `duplicate_tags`;

DROP TEMPORARY TABLE `_migration_tag_guard`;

UPDATE `tags` SET `name` = TRIM(`name`);

ALTER TABLE `tags`
  MODIFY COLUMN `name` varchar(64) NOT NULL,
  ADD UNIQUE KEY `uni_tags_name` (`name`);
