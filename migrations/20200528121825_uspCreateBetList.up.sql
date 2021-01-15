create or alter proc dbo.uspCreateBetList @TVP dbo.BetListType READONLY as
begin
    set nocount on

    MERGE dbo.BetList AS t
    USING @TVP s
    ON (t.Id = s.Id)

    WHEN MATCHED THEN
        UPDATE
        SET OrderType = s.OrderType,
            BetType = s.BetType,
            BetTypeDescription = s.BetTypeDescription,
            BetTypeTemplate = s.BetTypeTemplate,
            Sport = s.Sport,
            Placer = s.Placer,
            WantPrice = s.WantPrice,
            Price = s.Price,
            CcyRate = s.CcyRate,
            PlacementTime = s.PlacementTime,
            ExpiryTime = s.ExpiryTime,
            Closed = s.Closed,
            CloseReason = s.CloseReason,
            Status = s.Status,
            TakeStartingPrice = s.TakeStartingPrice,
            KeepOpenIr = s.KeepOpenIr,
            WantStake = s.WantStake,
            Stake = s.Stake,
            ProfitLoss = s.ProfitLoss,
            EventId = s.EventId,
            HomeId = s.HomeId,
            HomeTeam = s.HomeTeam,
            AwayId = s.AwayId,
            AwayTeam = s.AwayTeam,
            CompetitionId = s.CompetitionId,
            CompetitionName = s.CompetitionName,
            CompetitionCountry = s.CompetitionCountry,
            StartTime = s.StartTime,
            Date = s.Date,
            UpdatedAt =sysdatetimeoffset()

    WHEN NOT MATCHED THEN
        INSERT (Id, OrderType, BetType, BetTypeDescription, BetTypeTemplate, Sport, Placer, WantPrice, Price, CcyRate,
                PlacementTime, ExpiryTime, Closed, CloseReason, Status, TakeStartingPrice, KeepOpenIr, WantStake, Stake,
                ProfitLoss, EventId, HomeId, HomeTeam, AwayId, AwayTeam, CompetitionId, CompetitionName,
                CompetitionCountry, StartTime, Date)
        VALUES (s.Id, s.OrderType, s.BetType, s.BetTypeDescription, s.BetTypeTemplate, s.Sport, s.Placer, s.WantPrice,
                s.Price, s.CcyRate,
                s.PlacementTime, s.ExpiryTime, s.Closed, s.CloseReason, s.Status, s.TakeStartingPrice, s.KeepOpenIr,
                s.WantStake, s.Stake,
                s.ProfitLoss, s.EventId, s.HomeId, s.HomeTeam, s.AwayId, s.AwayTeam, s.CompetitionId, s.CompetitionName,
                s.CompetitionCountry, s.StartTime, s.Date);
end

