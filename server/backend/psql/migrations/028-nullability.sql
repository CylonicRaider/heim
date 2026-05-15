-- +migrate Up

ALTER TABLE room
    ALTER pk_nonce SET NOT NULL,
    ALTER pk_iv SET NOT NULL,
    ALTER pk_mac SET NOT NULL,
    ALTER encrypted_management_key SET NOT NULL,
    ALTER encrypted_private_key SET NOT NULL,
    ALTER public_key SET NOT NULL;

ALTER TABLE presence
    ALTER fact SET NOT NULL;

ALTER TABLE master_key
    ALTER encrypted_key SET NOT NULL,
    ALTER iv SET NOT NULL,
    ALTER nonce SET NOT NULL;

ALTER TABLE capability
    ALTER encrypted_private_data SET NOT NULL,
    ALTER public_data SET NOT NULL,
    ALTER nonce SET NOT NULL;

ALTER TABLE agent
    ALTER blessed SET NOT NULL;

-- +migrate Down

ALTER TABLE room
    ALTER pk_nonce DROP NOT NULL,
    ALTER pk_iv DROP NOT NULL,
    ALTER pk_mac DROP NOT NULL,
    ALTER encrypted_management_key DROP NOT NULL,
    ALTER encrypted_private_key DROP NOT NULL,
    ALTER public_key DROP NOT NULL;

ALTER TABLE presence
    ALTER fact DROP NOT NULL;

ALTER TABLE master_key
    ALTER encrypted_key DROP NOT NULL,
    ALTER iv DROP NOT NULL,
    ALTER nonce DROP NOT NULL;

ALTER TABLE capability
    ALTER encrypted_private_data DROP NOT NULL,
    ALTER public_data DROP NOT NULL,
    ALTER nonce DROP NOT NULL;

ALTER TABLE agent
    ALTER blessed DROP NOT NULL;
