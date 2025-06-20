-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS registration_otps(
    member_id INTEGER NOT NULL,
    member_reg_branch_id INTEGER NOT NULL,
    otp VARCHAR(6) NOT NULL,
    expired_at TEXT,
    PRIMARY KEY (member_id, member_reg_branch_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS registration_otps;
-- +goose StatementEnd
