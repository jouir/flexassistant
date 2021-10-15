-- Migrate from 1.2 to 1.3
-- SQLite doesn't support ALTER TABLE... ALTER TYPE
-- This migration file creates the new structures with "real" types
-- instead of "integer", move data to this new table then drop the old one.

-- miners

ALTER TABLE `miners` RENAME TO `miners_old`;

CREATE TABLE `miners` (
    `id` integer,
    `created_at` datetime,
    `updated_at` datetime,
    `deleted_at` datetime,
    `coin` text,
    `address` text NOT NULL UNIQUE,
    `balance` real,
    `last_payment_timestamp` integer,
    PRIMARY KEY (`id`)
);

INSERT INTO `miners` SELECT * FROM `miners_old`;

DROP TABLE `miners_old`;

CREATE INDEX `idx_miners_deleted_at` ON `miners`(`deleted_at`);

SELECT "Database migrated to 1.3"