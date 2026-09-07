-- 校内用户身份：username 继续作为登录账号和公开昵称，新增学号与真实姓名；
-- 邮箱改为数据库级必填且唯一；团队成员关系不再重复保存姓名备注。
--
-- 本迁移包含 DDL，MySQL 不保证整体事务回滚。执行前请先停止旧后端并备份数据库。
-- 如果历史用户存在空邮箱或重复邮箱，下面的迁移守卫会在任何 DDL 前终止；
-- 请先为这些账号补录真实、唯一的邮箱后重新执行。

CREATE TEMPORARY TABLE `_migration_school_identity_guard` (
  `invalid_count` bigint NOT NULL,
  CONSTRAINT `chk_school_identity_guard` CHECK (`invalid_count` = 0)
);

INSERT INTO `_migration_school_identity_guard` (`invalid_count`)
SELECT COUNT(*) FROM `users` WHERE `email` IS NULL OR CHAR_LENGTH(TRIM(`email`)) = 0;

TRUNCATE TABLE `_migration_school_identity_guard`;

INSERT INTO `_migration_school_identity_guard` (`invalid_count`)
SELECT COUNT(*)
FROM (
  SELECT `email`
  FROM `users`
  GROUP BY `email`
  HAVING COUNT(*) > 1
) AS `duplicate_emails`;

DROP TEMPORARY TABLE `_migration_school_identity_guard`;

ALTER TABLE `users`
  ADD COLUMN `student_no` varchar(32) DEFAULT NULL AFTER `username`,
  ADD COLUMN `real_name` varchar(64) NOT NULL DEFAULT '' AFTER `student_no`;

-- 优先沿用原团队备注中的真实姓名；同一用户存在多个备注时取非空最大值。
UPDATE `users` AS `u`
JOIN (
  SELECT `user_id`, MAX(NULLIF(TRIM(`remark`), '')) AS `real_name`
  FROM `team_members`
  GROUP BY `user_id`
) AS `tm` ON `tm`.`user_id` = `u`.`id`
SET `u`.`real_name` = `tm`.`real_name`
WHERE `tm`.`real_name` IS NOT NULL;

-- 没有团队备注的历史账号先用 username 兜底，管理员之后可补录真实姓名。
UPDATE `users`
SET `real_name` = `username`
WHERE CHAR_LENGTH(TRIM(`real_name`)) = 0;

ALTER TABLE `users`
  MODIFY COLUMN `real_name` varchar(64) NOT NULL,
  MODIFY COLUMN `email` varchar(191) NOT NULL,
  ADD UNIQUE KEY `uni_users_student_no` (`student_no`),
  ADD UNIQUE KEY `uni_users_email` (`email`),
  ADD CONSTRAINT `chk_users_email_nonblank` CHECK (CHAR_LENGTH(TRIM(`email`)) > 0),
  ADD CONSTRAINT `chk_users_real_name_nonblank` CHECK (CHAR_LENGTH(TRIM(`real_name`)) > 0);

ALTER TABLE `team_members`
  DROP COLUMN `remark`;
