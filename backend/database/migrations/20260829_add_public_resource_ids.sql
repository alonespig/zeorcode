-- 对外路由使用 8 位数字编号；原 id 继续作为内部主键和外键。
-- 历史数据用 0..89999999 上的双射完成无碰撞回填，新数据由应用使用 crypto/rand 生成。
ALTER TABLE `contests` ADD COLUMN `public_id` BIGINT NULL AFTER `id`;
ALTER TABLE `teams` ADD COLUMN `public_id` BIGINT NULL AFTER `id`;
ALTER TABLE `homeworks` ADD COLUMN `public_id` BIGINT NULL AFTER `id`;
ALTER TABLE `problem_sets` ADD COLUMN `public_id` BIGINT NULL AFTER `id`;

UPDATE `contests` SET `public_id` = 10000000 + MOD(`id` * 48271 + 1103, 90000000);
UPDATE `teams` SET `public_id` = 10000000 + MOD(`id` * 48271 + 2207, 90000000);
UPDATE `homeworks` SET `public_id` = 10000000 + MOD(`id` * 48271 + 3301, 90000000);
UPDATE `problem_sets` SET `public_id` = 10000000 + MOD(`id` * 48271 + 4409, 90000000);

ALTER TABLE `contests`
  MODIFY COLUMN `public_id` BIGINT NOT NULL,
  ADD UNIQUE KEY `idx_contests_public_id` (`public_id`);
ALTER TABLE `teams`
  MODIFY COLUMN `public_id` BIGINT NOT NULL,
  ADD UNIQUE KEY `idx_teams_public_id` (`public_id`);
ALTER TABLE `homeworks`
  MODIFY COLUMN `public_id` BIGINT NOT NULL,
  ADD UNIQUE KEY `idx_homeworks_public_id` (`public_id`);
ALTER TABLE `problem_sets`
  MODIFY COLUMN `public_id` BIGINT NOT NULL,
  ADD UNIQUE KEY `idx_problem_sets_public_id` (`public_id`);
