-- 提交对外使用 8 位随机数字编号；原 id 继续作为评测队列、Outbox 和测试点结果的内部关联键。
ALTER TABLE `submissions` ADD COLUMN `public_id` BIGINT NULL AFTER `id`;

-- 在当前小规模部署下，以内部主键做 0..89999999 上的无碰撞映射，回填历史提交。
-- 新提交由应用使用 crypto/rand 生成，数据库唯一索引负责最终兜底。
UPDATE `submissions`
SET `public_id` = 10000000 + MOD(`id` * 48271 + 5519, 90000000);

ALTER TABLE `submissions`
  MODIFY COLUMN `public_id` BIGINT NOT NULL,
  ADD UNIQUE KEY `idx_submissions_public_id` (`public_id`);
