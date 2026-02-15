-- Add column "dated" to table: "transaction"
ALTER TABLE `transaction` ADD COLUMN `dated` datetime NOT NULL DEFAULT '2026-02-15 00:00:00';
