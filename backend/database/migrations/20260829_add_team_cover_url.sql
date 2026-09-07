-- 团队封面由通用图片上传接口生成 URL；空字符串表示使用前端默认封面。
ALTER TABLE `teams`
  ADD COLUMN `cover_url` varchar(255) NOT NULL DEFAULT '' AFTER `name`;
