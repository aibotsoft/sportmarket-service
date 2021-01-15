package store

import (
	"context"
	"database/sql"
	"fmt"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/cache"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/sportmarket-service/pkg/token"
	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/dgraph-io/ristretto"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type Store struct {
	cfg   *config.Config
	log   *zap.SugaredLogger
	db    *sqlx.DB
	Cache *ristretto.Cache
}

func New(cfg *config.Config, log *zap.SugaredLogger, db *sqlx.DB) *Store {
	return &Store{log: log, db: db, Cache: cache.NewCache(cfg)}
}
func (s *Store) Close() {
	err := s.db.Close()
	if err != nil {
		s.log.Error(err)
	}
	s.Cache.Close()
}

func (s *Store) GetResults(ctx context.Context) (res []pb.BetResult, err error) {
	//return nil, nil
	rows, err := s.db.QueryxContext(ctx, "select * from dbo.vGetResults")
	if err != nil {
		return nil, errors.Wrap(err, "get_bet_results_error")
	}
	for rows.Next() {
		var r pb.BetResult
		var Price, Stake, WinLoss *float64
		var ApiBetId, ApiBetStatus *string
		err := rows.Scan(&r.SurebetId, &r.SideIndex, &r.BetId, &ApiBetId, &ApiBetStatus, &Price, &Stake, &WinLoss)
		if err != nil {
			s.log.Error(err)
			continue
		}
		if ApiBetId != nil {
			r.ApiBetId = *ApiBetId
		}
		if ApiBetStatus != nil {
			r.ApiBetStatus = *ApiBetStatus
		}
		if Price != nil {
			r.Price = *Price
		}
		if Stake != nil {
			r.Stake = *Stake
		}
		if WinLoss != nil {
			r.WinLoss = *WinLoss
		}
		res = append(res, r)
	}
	return res, nil
}

func (s *Store) LoadToken(ctx context.Context) (token token.Token, err error) {
	err = s.db.GetContext(ctx, &token, "select top 1 Session, CreatedAt, LastCheckAt from dbo.Auth order by CreatedAt desc ")
	return
}

func (s *Store) SaveBet(sb *pb.Surebet) error {
	side := sb.Members[0]
	_, err := s.db.Exec("dbo.uspSaveBet",
		sql.Named("SurebetId", sb.SurebetId),
		sql.Named("SideIndex", side.Num-1),

		sql.Named("BetId", side.ToBet.Id),
		sql.Named("TryCount", side.GetToBet().GetTryCount()),
		sql.Named("Status", side.GetBet().GetStatus()),
		sql.Named("StatusInfo", side.GetBet().GetStatusInfo()),
		sql.Named("Start", side.GetBet().GetStart()),
		sql.Named("Done", side.GetBet().GetDone()),
		sql.Named("Price", side.GetBet().GetPrice()),
		sql.Named("Stake", side.GetBet().GetStake()),
		sql.Named("ApiBetId", side.GetBet().GetApiBetId()),
	)
	if err != nil {
		return errors.Wrap(err, "uspSaveBet error")
	}
	return nil
}

func (s *Store) SaveCheck(sb *pb.Surebet) error {
	side := sb.Members[0]
	_, err := s.db.Exec("dbo.uspSaveSide",
		sql.Named("Id", sb.SurebetId),
		sql.Named("SideIndex", side.Num-1),

		sql.Named("SportName", side.SportName),
		sql.Named("SportId", side.SportId),
		sql.Named("LeagueName", side.LeagueName),
		sql.Named("LeagueId", side.LeagueId),
		sql.Named("Home", side.Home),
		sql.Named("HomeId", side.HomeId),
		sql.Named("Away", side.Away),
		sql.Named("AwayId", side.AwayId),
		sql.Named("MarketName", side.MarketName),
		sql.Named("MarketId", side.MarketId),
		sql.Named("Price", side.Price),
		sql.Named("Initiator", side.Initiator),
		sql.Named("Starts", side.Starts),
		sql.Named("EventId", side.EventId),

		sql.Named("CheckId", side.GetCheck().GetId()),
		sql.Named("AccountLogin", side.GetCheck().GetAccountLogin()),
		sql.Named("CheckPrice", side.GetCheck().GetPrice()),
		sql.Named("CheckStatus", side.GetCheck().GetStatus()),
		sql.Named("CountLine", side.GetCheck().GetCountLine()),
		sql.Named("CountEvent", side.GetCheck().GetCountEvent()),
		sql.Named("AmountEvent", side.GetCheck().GetAmountEvent()),
		sql.Named("AmountLine", side.GetCheck().GetAmountLine()),
		sql.Named("MinBet", side.GetCheck().GetMinBet()),
		sql.Named("MaxBet", side.GetCheck().GetMaxBet()),
		sql.Named("Balance", side.GetCheck().GetBalance()),
		sql.Named("Currency", side.GetCheck().GetCurrency()),
		sql.Named("CheckDone", side.GetCheck().GetDone()),

		sql.Named("CalcStatus", side.GetCheckCalc().GetStatus()),
		sql.Named("MaxStake", side.GetCheckCalc().GetMaxStake()),
		sql.Named("MinStake", side.GetCheckCalc().GetMinStake()),
		sql.Named("MaxWin", side.GetCheckCalc().GetMaxWin()),
		sql.Named("Stake", side.GetCheckCalc().GetStake()),
		sql.Named("Win", side.GetCheckCalc().GetWin()),
		sql.Named("IsFirst", side.GetCheckCalc().GetIsFirst()),
		sql.Named("ServiceName", side.ServiceName),
	)
	if err != nil {
		return errors.Wrapf(err, "uspSaveSide error")
	}
	return nil
}

type Stat struct {
	MarketName  string
	CountEvent  int64
	CountLine   int64
	AmountEvent int64
	AmountLine  int64
}

func (s *Store) GetStat(side *pb.SurebetSide) error {
	var stat []Stat
	err := s.db.Select(&stat, "dbo.uspCalcStat",
		sql.Named("EventId", side.EventId),
		sql.Named("ServiceName", side.ServiceName),
	)
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "uspCalcStat error")
	} else {
		for i := range stat {
			side.Check.AmountEvent = stat[i].AmountEvent
			side.Check.CountEvent = stat[i].CountEvent
			if stat[i].MarketName == side.MarketName {
				side.Check.CountLine = stat[i].CountLine
				side.Check.AmountLine = stat[i].AmountLine
				return nil
			}
		}
	}
	return nil
}

type Bet struct {
	Id                 int64
	OrderType          string
	BetType            string
	BetTypeDescription string
	BetTypeTemplate    string
	Sport              string
	Placer             string
	WantPrice          float64
	Price              float64
	CcyRate            float64
	PlacementTime      time.Time
	ExpiryTime         time.Time
	Closed             bool
	CloseReason        string
	Status             string
	TakeStartingPrice  bool
	KeepOpenIr         bool
	WantStake          float64
	Stake              float64
	ProfitLoss         *float64

	EventId            string
	HomeId             int64
	HomeTeam           string
	AwayId             int64
	AwayTeam           string
	CompetitionId      int64
	CompetitionName    string
	CompetitionCountry string
	StartTime          time.Time
	Date               string
}

func (s *Store) SaveBetList(ctx context.Context, bets []Bet) error {
	if len(bets) == 0 {
		return nil
	}
	tvp := mssql.TVP{TypeName: "BetListType", Value: bets}
	_, err := s.db.ExecContext(ctx, "dbo.uspCreateBetList", tvp)
	if err != nil {
		return errors.Wrap(err, "uspCreateBetList_error")
	}
	return nil
}

type League struct {
	Id      int64
	Name    string
	Country string
	Sport   string
}
type Team struct {
	Id   int64
	Name string
}

const saveLeagueQ = `
insert into dbo.League(Id, Name, Country, Sport)
select @Id, @Name, @Country, @Sport
where not exists(select 1 from dbo.League where Id = @Id)
`

func (s *Store) SaveLeague(league League) error {
	key := fmt.Sprintf("league:%d", league.Id)
	_, b := s.Cache.Get(key)
	if b {
		return nil
	}
	_, err := s.db.Exec(saveLeagueQ,
		sql.Named("Id", league.Id),
		sql.Named("Name", league.Name),
		sql.Named("Country", league.Country),
		sql.Named("Sport", league.Sport),
	)
	if err != nil {
		s.log.Error(err)
		return err
	}
	s.Cache.Set(key, struct{}{}, 1)
	return nil
}

const saveTeamQ = `
insert into dbo.Team(Id, Name)
select @Id, @Name
where not exists(select 1 from dbo.Team where Id = @Id)
`

func (s *Store) SaveTeam(team Team) error {
	key := fmt.Sprintf("team:%d", team.Id)
	_, b := s.Cache.Get(key)
	if b {
		return nil
	}
	_, err := s.db.Exec(saveTeamQ,
		sql.Named("Id", team.Id),
		sql.Named("Name", team.Name),
	)
	if err != nil {
		s.log.Error(err)
		return err
	}
	s.Cache.Set(key, struct{}{}, 1)
	return nil
}
func (s *Store) SaveEvent(event Event) error {
	key := fmt.Sprintf("event:%v", event.Id)
	_, b := s.Cache.Get(key)
	if b {
		return nil
	}
	_, err := s.db.Exec("dbo.uspSaveEvent",
		sql.Named("Id", event.Id),
		sql.Named("LeagueId", event.LeagueId),
		sql.Named("HomeId", event.HomeId),
		sql.Named("AwayId", event.AwayId),
		sql.Named("Starts", event.Starts),
	)
	if err != nil {
		s.log.Error(err)
		return err
	}
	s.Cache.Set(key, struct{}{}, 1)
	return nil
}

type Event struct {
	Id       string
	LeagueId int64
	HomeId   int64
	AwayId   int64
	Starts   time.Time
}

func (s *Store) GetStartTime(ctx context.Context, eventId string) (starts time.Time, err error) {
	key := fmt.Sprintf("starts:%v", eventId)
	got, b := s.Cache.Get(key)
	if b {
		return got.(time.Time), nil
	}
	err = s.db.GetContext(ctx, &starts, "select Starts from dbo.Event where Id=@p1", eventId)
	if err != nil {
		return starts, err
	}
	s.Cache.SetWithTTL(key, starts, 1, time.Minute*2)
	return
}

const GetEventByIdQ = `
select e.Id EventId, 
       e.LeagueId, 
       l.Name LeagueName,
       l.Sport,
       h.Name Home, 
       e.HomeId, 
       a.Name Away, 
       e.AwayId, 
       e.Starts 
from dbo.Event e 
         join dbo.League l on e.LeagueId = l.Id
		join dbo.Team h on h.Id = e.HomeId
		join dbo.Team a on a.Id = e.AwayId
where e.Id=@p1
`

func (s *Store) GetEventById(ctx context.Context, eventId string) (event UrlEvent, err error) {
	err = s.db.GetContext(ctx, &event, GetEventByIdQ, eventId)
	return
}

type UrlEvent struct {
	EventId      string
	Home         string
	HomeId       int64
	Away         string
	AwayId       int64
	Status       string
	Starts       time.Time
	Sport        string
	LeagueId     string
	LeagueName   string
	HomeScore    float64
	AwayScore    float64
	Avg          float64
	BetType      string
	PeriodNumber int
	Handicap     *float64
	IsLay        bool
}

const selectEventsByStartsQ = `
select e.Id   EventId,
       e.LeagueId,
       h.Name Home,
       e.HomeId,
       a.Name Away,
       e.AwayId,
       e.Starts,
       l.Name LeagueName,
       l.Sport
from dbo.Event e
         join dbo.League l on e.LeagueId = l.Id
		join dbo.Team h on h.Id = e.HomeId
		join dbo.Team a on a.Id = e.AwayId
where e.Starts>@p1 and e.Starts<@p2 and l.Sport = @p3
`

func (s *Store) SelectEventsByStarts(ctx context.Context, starts time.Time, spread time.Duration, sport string) (events []UrlEvent, err error) {
	err = s.db.SelectContext(ctx, &events, selectEventsByStartsQ, starts.Add(-spread), starts.Add(spread), sport)
	return
}
func (s *Store) SetVerifyWithTTL(key string, value interface{}, ttl time.Duration) bool {
	s.Cache.SetWithTTL(key, value, 1, ttl)
	for i := 0; i < 100; i++ {
		got, b := s.Cache.Get(key)
		if b {
			if got == value {
				return true
			} else {
				s.log.Info("got != value:", got, value)
				return false
			}
		}
		time.Sleep(time.Microsecond * 5)
	}
	return false
}

//const saveEventsQ = `
//insert into dbo.Event(Id, LeagueId, Home, HomeId, Away, AwayId, Status, Starts)
//select @Id, @LeagueId, @Home, @HomeId, @Away, @AwayId, @Status, @Starts
//where not exists(select 1 from dbo.Event where Id = @Id)
//`
//func (s *Store) SaveEvents(ctx context.Context, events []Event) error {
//
//	for i := range events {
//		_, b := s.Cache.Get(events[i].Id)
//		if !b {
//			s.log.Info(events[i])
//			_, err := s.db.ExecContext(ctx, saveEventsQ,
//				sql.Named("Id", events[i].Id),
//				sql.Named("LeagueId", events[i].LeagueId),
//				sql.Named("Home", events[i].Home),
//				sql.Named("HomeId", events[i].HomeId),
//				sql.Named("Away", events[i].Away),
//				sql.Named("AwayId", events[i].AwayId),
//				sql.Named("Status", events[i].Status),
//				sql.Named("Starts", events[i].Starts),
//			)
//			if err != nil {
//				s.log.Error(err)
//				continue
//			}
//			s.Cache.Set(events[i].Id, struct{}{}, 1)
//		}
//	}
//	return nil
//}

//func (s *Store) SaveLeagues(ctx context.Context, leagues []League) error {
//	for i := range leagues {
//		_, b := s.Cache.Get(leagues[i].Id)
//		if !b {
//			_, err := s.db.ExecContext(ctx, saveLeagueQ,
//				sql.Named("Id", leagues[i].Id),
//				sql.Named("Name", leagues[i].Name),
//				sql.Named("Country", leagues[i].Country),
//				sql.Named("Rank", leagues[i].Rank),
//			)
//			if err != nil {
//				s.log.Error(err)
//				continue
//			}
//			s.Cache.Set(leagues[i].Id, struct{}{}, 1)
//		}
//	}
//	return nil
//}

//type Event struct {
//	Id       string
//	LeagueId int64
//	Home     string
//	HomeId   int64
//	Away     string
//	AwayId   int64
//	Status   string
//	Starts   time.Time
//	Sport    string
//}
//
//func (s *Store) SaveEvents(ctx context.Context, events []Event) error {
//	tvpType := mssql.TVP{TypeName: "EventType", Value: events}
//	_, err := s.db.ExecContext(ctx, "dbo.uspSaveEvents", tvpType)
//	if err != nil {
//		return errors.Wrap(err, "uspSaveEvents error")
//	}
//	return nil
//}
