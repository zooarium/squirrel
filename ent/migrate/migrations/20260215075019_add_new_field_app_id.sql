-- Add column "app_id" to table: "category"
ALTER TABLE `category` ADD COLUMN `app_id` integer NOT NULL;
-- Add column "app_id" to table: "transaction"
ALTER TABLE `transaction` ADD COLUMN `app_id` integer NOT NULL;
