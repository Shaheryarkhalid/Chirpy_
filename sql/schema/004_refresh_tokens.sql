-- +goose UP

CREATE TABLE refresh_tokens(
    token text primary key, 
    created_at timestamp not null,
    updated_at timestamp not null, 
    user_id uuid not null, 
    expires_at timestamp not null, 
    revoked_at timestamp,  
    constraint fk_user_id foreign key (user_id) references users(id) on delete cascade
);

-- +goose DOWN
DROP TABLE  refresh_tokens;
