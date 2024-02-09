-- +migrate Up
-- give account names more meaning than "just a random string"

CREATE UNIQUE INDEX IF NOT EXISTS account_name_unique
    ON account(LOWER(REPLACE(name, ' ', '')))
    WHERE name <> '';

-- +migrate Down

DROP INDEX IF EXISTS account_name_unique;
