-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_categories" table
CREATE TABLE `new_categories` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `user_id` integer NOT NULL, `name` text NOT NULL, `status` integer NOT NULL DEFAULT (1));
-- Copy rows from old table "categories" to new temporary table "new_categories"
INSERT INTO `new_categories` (`id`, `created_at`, `updated_at`, `name`) SELECT `id`, `created_at`, `updated_at`, `name` FROM `categories`;
-- Drop "categories" table after copying rows
DROP TABLE `categories`;
-- Rename temporary table "new_categories" to "categories"
ALTER TABLE `new_categories` RENAME TO `categories`;
-- Create "transactions" table
CREATE TABLE `transactions` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `user_id` integer NOT NULL, `amount` real NOT NULL, `type` text NOT NULL, `category_id` integer NULL, CONSTRAINT `transactions_categories_transactions` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE SET NULL);
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
