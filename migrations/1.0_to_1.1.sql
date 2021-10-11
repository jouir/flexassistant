-- Migrate from 1.0 to 1.1
-- SQLite doesn't support ALTER TABLE... ALTER TYPE
-- This migration file creates the new structures with "integer" types
-- instead of "real", move data to this new table then drop the old one.

-- miners

ALTER TABLE `miners` RENAME TO `miners_old`;

CREATE TABLE `miners` (
    `id` integer,
    `created_at` datetime,
    `updated_at` datetime,
    `deleted_at` datetime,
    `coin` text,
    `address` text NOT NULL UNIQUE,
    `balance` integer,
    `last_payment_timestamp` integer,
    PRIMARY KEY (`id`)
);

INSERT INTO `miners` SELECT * FROM `miners_old`;

DROP TABLE `miners_old`;

CREATE INDEX `idx_miners_deleted_at` ON `miners`(`deleted_at`);


-- pools

ALTER TABLE `pools` RENAME TO `pools_old`;

CREATE TABLE `pools` (
    `id` integer,
    `created_at` datetime,
    `updated_at` datetime,
    `deleted_at` datetime,
    `coin` text NOT NULL UNIQUE,
    `last_block_number` integer,
    PRIMARY KEY (`id`)
);

INSERT INTO `pools` SELECT * FROM `pools_old`;

DROP TABLE `pools_old`;

CREATE INDEX `idx_pools_deleted_at` ON `pools`(`deleted_at`);

SELECT "Database migrated to 1.1"