ALTER TABLE `contests`
  ADD COLUMN `archived` tinyint(1) NOT NULL DEFAULT '0' AFTER `settled`,
  ADD KEY `idx_contests_archived_start_time` (`archived`, `start_time`);
