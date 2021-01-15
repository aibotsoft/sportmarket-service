create or alter view dbo.vGetResults as
select top 1200 b.SurebetId,
                b.SideIndex,
               b.BetId,
               b.ApiBetId,
               l.Status                                                                                           ApiBetStatus,
               l.Price                                                                                            Price,
               cast(l.Stake / (select top 1 ss.Currency from Side ss where ss.Id = b.SurebetId) as decimal(9, 2)) Stake,
               cast(l.ProfitLoss /
                    (select top 1 ss.Currency from Side ss where ss.Id = b.SurebetId) as decimal(9, 2))           WinLoss
from Bet b
         left join dbo.BetList l on b.ApiBetId = l.Id
where b.ApiBetId > 0
  and l.Status is not null
order by b.SurebetId desc
