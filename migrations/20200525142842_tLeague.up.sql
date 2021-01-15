create table dbo.League
(
    Id        int                                        not null,
    Name      varchar(1000)                              not null,
    Country   varchar(300),
    Rank      int,
    Sport     varchar(300),
    CreatedAt datetimeoffset default sysdatetimeoffset() not null,
    UpdatedAt datetimeoffset default sysdatetimeoffset() not null,
    constraint PK_League primary key (Id),
)