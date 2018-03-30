create table item
(
    item_id serial not null,
    id character varying not null, 
    type character varying not null,
    raw character varying not null,
    source jsonb not null,
    constraint item_pkey primary key (item_id)
);