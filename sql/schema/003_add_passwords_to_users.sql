-- +goose UP

alter table users add COLUMN hashed_password text not null default 'unset';

-- +goose DOWN

alter table users drop COLUMN hashed_password;
