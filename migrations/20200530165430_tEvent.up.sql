create table dbo.Event
(
    Id        varchar(300)                                  not null,
    LeagueId  int                                           not null,
    Home      varchar(500),
    HomeId    int,
    Away      varchar(500),
    AwayId    int,
    Status    varchar(300),
    Sport     varchar(300),
    Starts    datetimeoffset(0),
    CreatedAt datetimeoffset(0) default sysdatetimeoffset() not null,
    UpdatedAt datetimeoffset(0) default sysdatetimeoffset() not null,
    constraint PK_Event primary key (Id),
);
create type dbo.EventType as table
(
    Id       varchar(300) not null,
    LeagueId int          not null,
    Home     varchar(500),
    HomeId   int,
    Away     varchar(500),
    AwayId   int,
    Status   varchar(300),
    Starts   datetimeoffset(0),
    Sport    varchar(300),
    primary key (Id)
);