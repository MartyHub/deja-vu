create table country
(
    name                     text not null,
    alpha2                   text not null,
    alpha3                   text not null,
    country_code             text not null,
    iso_3166_2               text not null,
    region                   text,
    sub_region               text,
    intermediate_region      text,
    region_code              text,
    sub_region_code          text,
    intermediate_region_code text,
    constraint country_pk primary key (name)
);
