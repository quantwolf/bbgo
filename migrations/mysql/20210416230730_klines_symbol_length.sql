-- +up
ALTER TABLE `klines`
MODIFY COLUMN `symbol` VARCHAR(10) NOT NULL;

ALTER TABLE `okex_klines`
MODIFY COLUMN `symbol` VARCHAR(10) NOT NULL;

ALTER TABLE `binance_klines`
MODIFY COLUMN `symbol` VARCHAR(10) NOT NULL;

ALTER TABLE `max_klines`
MODIFY COLUMN `symbol` VARCHAR(10) NOT NULL;

-- +down
ALTER TABLE `klines`
MODIFY COLUMN `symbol` VARCHAR(7) NOT NULL;

ALTER TABLE `okex_klines`
MODIFY COLUMN `symbol` VARCHAR(7) NOT NULL;

ALTER TABLE `binance_klines`
MODIFY COLUMN `symbol` VARCHAR(7) NOT NULL;

ALTER TABLE `max_klines`
MODIFY COLUMN `symbol` VARCHAR(7) NOT NULL;
