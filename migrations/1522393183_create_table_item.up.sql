create table item
(
    item_id serial not null,
    abstract character varying, 
    id character varying, 
    type character varying not null,
    source character varying not null,
    raw jsonb not null,
    constraint item_pkey primary key (item_id)
);