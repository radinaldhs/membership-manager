-- +goose Up
-- +goose StatementBegin
ALTER TABLE registration_otps 
    ADD COLUMN next_regeneration TEXT NOT NULL DEFAULT '0001-01-01T00:00:00Z';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE registration_otps
    DROP COLUMN next_regeneration;
-- +goose StatementEnd
