-- Existing notifications have NULL keys and are unaffected by the unique index.
-- Leave content unchanged: deployments may already use either VARCHAR(255) or VARCHAR(2000).
ALTER TABLE `notifications`
  ADD COLUMN `dedupe_key` varchar(96) CHARACTER SET ascii COLLATE ascii_bin DEFAULT NULL,
  ADD UNIQUE KEY `idx_notifications_dedupe_key` (`dedupe_key`);
