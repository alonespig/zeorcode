-- 管理员权限表：
-- users.role: 0 普通用户 / 1 管理员 / 2 超级管理员。
-- 超级管理员默认拥有全部权限，不写 admin_permissions。
CREATE TABLE IF NOT EXISTS `admin_permissions` (
  `user_id` bigint NOT NULL,
  `permission` varchar(64) NOT NULL,
  `created_at` datetime(3) NOT NULL,
  PRIMARY KEY (`user_id`,`permission`),
  KEY `idx_admin_permissions_permission` (`permission`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 升级前 role=1 的账号提升为超级管理员，避免升级后无人能进入后台。
-- 仅在系统中还没有任何超级管理员时执行：迁移可重复运行，不会把后来创建的
-- 受限 admin 再次提升为超管。升级完成后可在用户管理里再逐项分配权限。
UPDATE `users` AS u
SET u.`role` = 2
WHERE u.`role` = 1
  AND NOT EXISTS (
    SELECT 1
    FROM (SELECT `id` FROM `users` WHERE `role` = 2 LIMIT 1) AS supers
  );
