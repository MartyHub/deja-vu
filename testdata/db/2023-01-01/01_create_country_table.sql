create table country
(
    name                     varchar(64) not null,
    alpha2                   char(2)     not null,
    alpha3                   char(3)     not null,
    country_code             char(3)     not null,
    iso_3166_2               char(13)    not null,
    region                   varchar(8),
    sub_region               varchar(32),
    intermediate_region      varchar(16),
    region_code              char(3),
    sub_region_code          char(3),
    intermediate_region_code char(3),
    primary key (name)
);
