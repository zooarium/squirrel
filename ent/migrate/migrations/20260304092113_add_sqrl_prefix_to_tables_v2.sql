-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_sqrl_transaction" table
CREATE TABLE `new_sqrl_transaction` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `app_id` integer NOT NULL, `user_id` integer NOT NULL, `amount` real NOT NULL, `type` text NOT NULL, `recurring` integer NOT NULL DEFAULT (0), `dated` datetime NOT NULL, `category_id` integer NULL, CONSTRAINT `sqrl_transaction_sqrl_category_transactions` FOREIGN KEY (`category_id`) REFERENCES `sqrl_category` (`id`) ON DELETE SET NULL);
-- Copy rows from old table "sqrl_transaction" to new temporary table "new_sqrl_transaction"
INSERT INTO `new_sqrl_transaction` (`id`, `created_at`, `updated_at`, `app_id`, `user_id`, `amount`, `type`, `recurring`, `dated`, `category_id`) SELECT `id`, `created_at`, `updated_at`, `app_id`, `user_id`, `amount`, `type`, `recurring`, `dated`, `category_id` FROM `sqrl_transaction`;
-- Drop "sqrl_transaction" table after copying rows
DROP TABLE `sqrl_transaction`;
-- Rename temporary table "new_sqrl_transaction" to "sqrl_transaction"
ALTER TABLE `new_sqrl_transaction` RENAME TO `sqrl_transaction`;
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
