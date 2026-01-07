-- +goose UP

CREATE TABLE chirps(id uuid primary key,  created_at timestamp not null, updated_at timestamp not null, body text not null, user_id uuid not null, constraint fk_user_id foreign key (user_id)  references users(id) on delete cascade);
