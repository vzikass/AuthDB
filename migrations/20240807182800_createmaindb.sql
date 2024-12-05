-- +goose Up
-- +goose StatementBegin

-- Users table
create table if not exists users (
    id bigserial primary key,
    username varchar(255) unique not null,
    email varchar(255) unique not null,
    password varchar(255) not null,
    role varchar(50) not null default 'user',
    created_at timestamp default CURRENT_TIMESTAMP
);

-- Session table for GoAdmin
create table goadmin_session (
    id bigserial primary key,
    sid varchar(255) not null unique,
    values text not null,
    created_at timestamp default CURRENT_TIMESTAMP,
    updated_at timestamp default CURRENT_TIMESTAMP
);

-- Site Settings Table for GoAdmin
create table goadmin_site (
    id bigserial primary key,
    key varchar(255) not null,
    value varchar(255) not null,
    state boolean default TRUE,
    created_at timestamp default CURRENT_TIMESTAMP
);

-- Table for GoAdmin user roles
create table goadmin_roles (
    id serial primary key,
    name varchar(50) not null,
    slug varchar(50) not null,
    created_at timestamp default CURRENT_TIMESTAMP,
    updated_at timestamp default CURRENT_TIMESTAMP
);

-- insert into goadmin_roles (slug, name) values ('admin', 'admin');

-- Table for GoAdmin users
create table if not exists goadmin_users (
    id bigserial primary key,
    username varchar(255) unique not null,
    password varchar(255) not null,
    role_id int references goadmin_roles(id), 
    created_at timestamp default CURRENT_TIMESTAMP,
    updated_at timestamp default CURRENT_TIMESTAMP
);

-- Table linking users and roles
create table goadmin_role_users (
    id serial primary key,
    role_id int not null references goadmin_roles(id) on delete cascade,
    user_id int not null references goadmin_users(id) on delete cascade
); 

-- Table for storing authorizations
create table if not exists goadmin_permissions (
    id bigserial primary key,
    name varchar(255),
    slug varchar(255) not null, 
    permission varchar(255) not null,
    http_method varchar(10),
    http_path varchar(255),   
    created_at timestamp default CURRENT_TIMESTAMP,
    updated_at timestamp default CURRENT_TIMESTAMP
);

create table if not exists goadmin_user_permissions (
    id serial primary key,
    user_id int not null references goadmin_users(id) on delete cascade,
    permission_id int not null references goadmin_permissions(id) on delete cascade,
    created_at timestamp default CURRENT_TIMESTAMP
);

-- Table linking roles and permissions
create table goadmin_role_permissions (
    id serial primary key,
    role_id int not null references goadmin_roles(id) on delete cascade,
    permission_id int not null references goadmin_permissions(id) on delete cascade,
    created_at timestamp default CURRENT_TIMESTAMP
);

create table if not exists goadmin_menu (
    id serial primary key,
    parent_id int,
    title varchar(255) not null,
    icon varchar(255),
    uri varchar(255),
    created_at timestamp default CURRENT_TIMESTAMP,
    updated_at timestamp default CURRENT_TIMESTAMP
);

create table if not exists goadmin_role_menu (
    role_id int not null references goadmin_roles(id) on delete cascade,
    menu_id int not null references goadmin_menu(id) on delete cascade,
    created_at timestamp default CURRENT_TIMESTAMP
)
-- +goose StatementEnd
