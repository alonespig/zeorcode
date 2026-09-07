-- 加速作业题目列表的提交通过率统计：
-- WHERE homework_id = ? AND problem_id IN (?) GROUP BY problem_id。
-- submissions 已保存作业归属，不需要额外的统计表。
ALTER TABLE `submissions`
  ADD KEY `idx_submission_homework_problem_status` (`homework_id`, `problem_id`, `status`);
