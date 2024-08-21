create table if not exists users (
    id serial primary key,
    login varchar(200) unique not null,
    email varchar(200) unique not null,
    password varchar(200) not null
);



