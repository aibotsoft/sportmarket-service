create or alter proc dbo.uspSaveEvent @Id varchar(300),
                                      @LeagueId int,
                                      @HomeId int,
                                      @AwayId int,
                                      @Starts datetimeoffset
as
begin
    set nocount on
    MERGE dbo.Event AS t
    USING (select @Id) s (Id)
    ON (t.Id = s.Id)

    WHEN MATCHED THEN
        UPDATE
        SET LeagueId  = @LeagueId,
            HomeId    = @HomeId,
            AwayId    = @AwayId,
            Starts    = @Starts,
            UpdatedAt = sysdatetimeoffset()

    WHEN NOT MATCHED THEN
        INSERT (Id, LeagueId, HomeId, AwayId, Starts)
        VALUES (s.Id, @LeagueId, @HomeId, @AwayId, @Starts);
end