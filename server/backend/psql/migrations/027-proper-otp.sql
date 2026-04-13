-- +migrate Up
-- An OTP may only be used once. The validation flag is converted to a last-use
-- counter, for which mononotic increase can be validated. Existing validation
-- records are dropped as a precaution.

ALTER TABLE otp
    DROP validated,
    ADD last_validated bigint DEFAULT 0; -- 0 is invalid

-- +migrate Down
-- validated could be reconstructed from last_validated, but revalidating an
-- OTP ought to be trivial.

ALTER TABLE otp
    ADD validated bool DEFAULT false,
    DROP last_validated;
