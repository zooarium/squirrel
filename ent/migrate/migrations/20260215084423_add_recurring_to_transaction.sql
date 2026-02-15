-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_transaction" table
CREATE TABLE `new_transaction` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `app_id` integer NOT NULL, `user_id` integer NOT NULL, `amount` real NOT NULL, `type` text NOT NULL, `recurring` integer NOT NULL DEFAULT (0), `category_id` integer NULL, CONSTRAINT `transaction_category_transactions` FOREIGN KEY (`category_id`) REFERENCES `category` (`id`) ON DELETE SET NULL);
-- Copy rows from old table "transaction" to new temporary table "new_transaction"
INSERT INTO `new_transaction` (`id`, `created_at`, `updated_at`, `app_id`, `user_id`, `amount`, `type`, `category_id`) SELECT `id`, `created_at`, `updated_at`, `app_id`, `user_id`, `amount`, `type`, `category_id` FROM `transaction`;
-- Drop "transaction" table after copying rows
DROP TABLE `transaction`;
-- Rename temporary table "new_transaction" to "transaction"
ALTER TABLE `new_transaction` RENAME TO `transaction`;
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
