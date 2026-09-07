-- 后台批量导入用户时允许暂不绑定邮箱。
-- MySQL 唯一索引允许存在多个 NULL，因此非空邮箱仍保持唯一。

ALTER TABLE `users`
  DROP CHECK `chk_users_email_nonblank`,
  MODIFY COLUMN `email` varchar(191) DEFAULT NULL;
