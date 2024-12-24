ALTER TABLE `task_results` ADD COLUMN `status` TEXT;

-- Fill column status based on previously recorded results.
DELETE FROM `task_results` WHERE `result` = 5; -- ResultNoChanges
DELETE FROM `task_results` WHERE `result` = 12; -- ResultNoMatch
DELETE FROM `task_results` WHERE `result` = 13; -- ResultSkip
UPDATE `task_results` SET `status` = 'error' WHERE `result` = 0;
UPDATE `task_results` SET `status` = 'closed' WHERE `result` = 7 OR `result` = 8;
UPDATE `task_results` SET `status` = 'merged' WHERE `result` = 9 OR `result` = 10;
UPDATE `task_results` SET `status` = 'open' WHERE `result` NOT IN (0, 7, 8, 9, 10);
