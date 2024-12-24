CREATE TABLE IF NOT EXISTS `runs` (
  `error` text,
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `finished_at` datetime,
  `reason` integer,
  `repository_names` text,
  `schedule_after` datetime,
  `started_at` datetime,
  `status` integer,
  `task_name` text
);
