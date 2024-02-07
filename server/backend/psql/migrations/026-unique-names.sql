-- +migrate Up
-- give account names more meaning than "just a random string"

ALTER TABLE account
    ADD CONSTRAINT account_name_no_space
    CHECK (name !~ '\s');

CREATE UNIQUE INDEX IF NOT EXISTS account_name_unique
    ON account(LOWER(name))
    WHERE name <> '';

-- +migrate Down

DROP INDEX IF EXISTS account_name_unique;
ALTER TABLE account DROP CONSTRAINT IF EXISTS account_name_no_space;
