create table item
(
    item_id serial not null,
    id character varying, 
    type character varying,
    name character varying,
    source character varying,
    constraint item_pkey primary key (item_id)
);