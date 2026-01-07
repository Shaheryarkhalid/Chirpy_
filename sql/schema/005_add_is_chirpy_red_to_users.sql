-- +goose UP
alter table users add  COLUMN is_chirpy_red boolean not null default false;

-- +goose DOWN
alter table users drop COLUMN is_chirpy_red; 
