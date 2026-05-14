-- +migrate Up

ALTER TABLE master_key
    ALTER encrypted_key SET NOT NULL,
    ALTER iv SET NOT NULL,
    ALTER nonce SET NOT NULL;

ALTER TABLE capability
    ALTER encrypted_private_data SET NOT NULL,
    ALTER public_data SET NOT NULL,
    ALTER nonce SET NOT NULL;

-- +migrate Down

ALTER TABLE master_key
    ALTER encrypted_key DROP NOT NULL,
    ALTER iv DROP NOT NULL,
    ALTER nonce DROP NOT NULL;

ALTER TABLE capability
    ALTER encrypted_private_data DROP NOT NULL,
    ALTER public_data DROP NOT NULL,
    ALTER nonce DROP NOT NULL;
