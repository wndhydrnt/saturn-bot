ALTER TABLE `task_results` ADD COLUMN `status` TEXT;

-- Fill column status based on previously recorded results.
UPDATE `task_results` SET `status` = 'error' WHERE `result` = 0;
UPDATE `task_results` SET `status` = 'closed' WHERE `result` = 7 OR `result` = 8;
UPDATE `task_results` SET `status` = 'merged' WHERE `result` = 9 OR `result` = 10;
UPDATE `task_results` SET `status` = 'open' WHERE `status` is null;
