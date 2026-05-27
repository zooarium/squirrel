-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_sqrl_category" table
CREATE TABLE `new_sqrl_category` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `app_id` integer NOT NULL, `user_id` integer NOT NULL, `name` text NOT NULL, `status` integer NOT NULL DEFAULT (1), `division_id` integer NULL, CONSTRAINT `sqrl_category_sqrl_division_categories` FOREIGN KEY (`division_id`) REFERENCES `sqrl_division` (`id`) ON DELETE SET NULL);
-- Copy rows from old table "sqrl_category" to new temporary table "new_sqrl_category"
INSERT INTO `new_sqrl_category` (`id`, `created_at`, `updated_at`, `app_id`, `user_id`, `name`, `status`) SELECT `id`, `created_at`, `updated_at`, `app_id`, `user_id`, `name`, `status` FROM `sqrl_category`;
-- Drop "sqrl_category" table after copying rows
DROP TABLE `sqrl_category`;
-- Rename temporary table "new_sqrl_category" to "sqrl_category"
ALTER TABLE `new_sqrl_category` RENAME TO `sqrl_category`;
-- Create "sqrl_division" table
CREATE TABLE `sqrl_division` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `app_id` integer NOT NULL, `name` text NOT NULL, `path` text NOT NULL, `depth` integer NOT NULL DEFAULT (0), `status` integer NOT NULL DEFAULT (1), `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `parent_id` integer NULL, CONSTRAINT `sqrl_division_sqrl_division_children` FOREIGN KEY (`parent_id`) REFERENCES `sqrl_division` (`id`) ON DELETE SET NULL);
-- Create index "division_path" to table: "sqrl_division"
CREATE INDEX `division_path` ON `sqrl_division` (`path`);
-- Create index "division_app_id_parent_id" to table: "sqrl_division"
CREATE INDEX `division_app_id_parent_id` ON `sqrl_division` (`app_id`, `parent_id`);
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
