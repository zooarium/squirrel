-- Create "category" table
CREATE TABLE `category` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `user_id` integer NOT NULL, `name` text NOT NULL, `status` integer NOT NULL DEFAULT (1));
-- Create "transaction" table
CREATE TABLE `transaction` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `user_id` integer NOT NULL, `amount` real NOT NULL, `type` text NOT NULL, `category_id` integer NULL, CONSTRAINT `transaction_category_transactions` FOREIGN KEY (`category_id`) REFERENCES `category` (`id`) ON DELETE SET NULL);
