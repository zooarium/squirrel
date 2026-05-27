-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Add column "division_id" to table: "sqrl_transaction"
ALTER TABLE `sqrl_transaction` ADD COLUMN `division_id` integer NULL;
-- Create "new_sqrl_category" table
CREATE TABLE `new_sqrl_category` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `app_id` integer NOT NULL, `user_id` integer NOT NULL, `division_id` integer NULL, `name` text NOT NULL, `status` integer NOT NULL DEFAULT (1));
-- Copy rows from old table "sqrl_category" to new temporary table "new_sqrl_category"
INSERT INTO `new_sqrl_category` (`id`, `created_at`, `updated_at`, `app_id`, `user_id`, `division_id`, `name`, `status`) SELECT `id`, `created_at`, `updated_at`, `app_id`, `user_id`, `division_id`, `name`, `status` FROM `sqrl_category`;
-- Drop "sqrl_category" table after copying rows
DROP TABLE `sqrl_category`;
-- Rename temporary table "new_sqrl_category" to "sqrl_category"
ALTER TABLE `new_sqrl_category` RENAME TO `sqrl_category`;
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
