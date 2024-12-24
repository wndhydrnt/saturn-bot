CREATE TABLE IF NOT EXISTS `task_results` (
  `created_at` datetime,
  `error` text,
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `repository_name` text,
  `result` integer,
  `run_id` integer,
  `task_name` text
);
