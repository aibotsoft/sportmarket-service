create or alter proc dbo.uspCalcStat @EventId varchar(300), @ServiceName varchar(100)
as
begin
    set nocount on;
    select s.MarketName,
           count(s.EventId) over ( )                                   CountEvent,
           cast(sum(b.Stake) over ( ) as int)                          AmountEvent,
           count(s.MarketName) over ( partition by s.MarketName)       CountLine,
           cast(sum(b.Stake) over ( partition by s.MarketName) as int) AmountLine
    from dbo.Bet b
             join dbo.Side s on s.Id = b.SurebetId and b.SideIndex=s.SideIndex
    where EventId = @EventId and b.Status = 'Ok' and s.ServiceName = @ServiceName
end

-- 2020-07-31,245,289
--     Х
-- exec dbo.uspCalcStat '2020-07-31,245,289', 'Против X', 'SportMarket'
--
-- select s.MarketName,
--        count(s.EventId) over ( )                                   CountEvent,
--        cast(sum(b.Stake) over ( ) as int)                          AmountEvent,
--        count(s.MarketName) over ( partition by s.MarketName)       CountLine,
--        cast(sum(b.Stake) over ( partition by s.MarketName) as int) AmountLine
-- from dbo.Bet b
--          join dbo.Side s on s.Id = b.SurebetId and b.SideIndex=s.SideIndex
-- where EventId = '2020-07-31,245,289'
--   and b.Status = 'Ok' and s.ServiceName = 'Matchbook'