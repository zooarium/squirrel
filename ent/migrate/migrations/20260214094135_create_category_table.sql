-- Create "categories" table
CREATE TABLE `categories` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `name` text NOT NULL);
-- Create index "categories_name_key" to table: "categories"
CREATE UNIQUE INDEX `categories_name_key` ON `categories` (`name`);
