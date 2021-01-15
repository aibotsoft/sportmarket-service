create or alter proc dbo.uspSaveEvents @TVP dbo.EventType READONLY as
begin
    set nocount on

    MERGE dbo.Event AS t
    USING @TVP s
    ON (t.Id = s.Id)

    WHEN MATCHED THEN
        UPDATE
        SET LeagueId  = s.LeagueId,
            Home      = s.Home,
            HomeId    = s.HomeId,
            Away      = s.Away,
            AwayId    = s.AwayId,
            Starts    = s.Starts,
            Status    = s.Status,
            Sport    = s.Sport,
            UpdatedAt = sysdatetimeoffset()

    WHEN NOT MATCHED THEN
        INSERT (Id,
                LeagueId,
                Home,
                HomeId,
                Away,
                AwayId,
                Status,
                Starts,
                Sport)
        values (s.Id,
                s.LeagueId,
                s.Home,
                s.HomeId,
                s.Away,
                s.AwayId,
                s.Status,
                s.Starts,
                s.Sport);
end
