-- 保存本地判题失败时的原始编译器输出，供提交者和管理员在评测详情页查看。
-- 重判会先清空旧输出，成功编译后的提交保持为空。

ALTER TABLE `submissions`
  ADD COLUMN `compile_output` mediumtext NULL AFTER `memory_used`;
