package collector

import (
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/sportmarket-service/pkg/client"
	"github.com/aibotsoft/sportmarket-service/pkg/store"
	"github.com/aibotsoft/sportmarket-service/services/auth"
	"go.uber.org/zap"
	"time"
)

type Collector struct {
	cfg    *config.Config
	log    *zap.SugaredLogger
	store  *store.Store
	client *client.Client
	auth   *auth.Auth
}

const collectPeriod = time.Second * 90

func New(cfg *config.Config, log *zap.SugaredLogger, store *store.Store, auth *auth.Auth) *Collector {
	return &Collector{cfg: cfg, log: log, store: store, client: client.New(cfg, log), auth: auth}
}

//
//func (c *Collector) CollectNowJob() {
//	for  {
//		if shared.NeedCollect {
//			c.log.Info("got_request_to_collect_now")
//			shared.NeedCollect = false
//			_, b := c.store.Cache.Get("NeedCollect")
//			if b {
//				time.Sleep(time.Second)
//				continue
//			}
//			c.store.Cache.SetWithTTL("NeedCollect", true, 1, time.Second*30)
//			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//			err := c.CollectRound(ctx)
//			cancel()
//			if err != nil {
//				c.log.Info(err)
//			}
//			c.log.Info("collect_now_done")
//		}
//		time.Sleep(time.Second)
//	}
//}
//func (c *Collector) CollectJob() {
//	go c.CollectNowJob()
//	for {
//		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
//		err := c.CollectRound(ctx)
//		cancel()
//		if err != nil {
//			c.log.Error(err)
//		}
//		time.Sleep(collectPeriod)
//	}
//}
//func (c *Collector) CollectRound(ctx context.Context) error {
//	authCtx, err := c.auth.Auth(ctx)
//	if err != nil {
//		return err
//	}
//	resp, err := c.client.Events(authCtx)
//	if err != nil {
//		return err
//	}
//	if resp.GetStatus() != "ok" {
//		return errors.Errorf("response_status_not_ok")
//	}
//	var leagues []store.League
//	var events []store.Event
//	for leagueIdStr, league := range resp.GetData() {
//		leagueId, err := strconv.ParseInt(leagueIdStr, 10, 64)
//		if err != nil {
//			c.log.Error(err)
//			continue
//		}
//		leagues = append(leagues, store.League{
//			Id:      leagueId,
//			Name:    league.GetName(),
//			Country: league.GetCountry(),
//			Rank:    league.GetRank(),
//		})
//		//c.log.Infow("", leagueIdStr, league.Name, "country", league.Rank)
//		for eventIdStr, event := range league.GetEvents() {
//			eventSplit := strings.Split(eventIdStr, ",")
//			if len(eventSplit) != 3 {
//				c.log.Info("split_event_error: ", eventIdStr)
//				continue
//			}
//			homeId, _ := strconv.ParseInt(eventSplit[1], 10, 64)
//			awayId, _ := strconv.ParseInt(eventSplit[2], 10, 64)
//			var sport string
//			if len(event.GetSports()) > 0 {
//				sport = event.GetSports()[0]
//				if sport == "fb_ht" {
//					sport = "fb"
//				} else if sport == "fb_htft" {
//					sport = "fb"
//				}else if sport == "fb_corn" {
//					sport = "fb"
//				}else if sport == "fb_corn_ht" {
//					sport = "fb"
//				}else if sport == "fb_et" {
//					sport = "fb"
//				}else if sport == "basket_ht" {
//					sport = "basket"
//				}
//			}
//
//			events = append(events, store.Event{
//				Id:       eventIdStr,
//				LeagueId: leagueId,
//				Home:     event.GetHome(),
//				HomeId:   homeId,
//				Away:     event.GetAway(),
//				AwayId:   awayId,
//				Status:   event.GetIrStatus(),
//				Starts:   event.GetStartTime(),
//				Sport:    sport,
//			})
//			//c.log.Infow("", eventIdStr, event)
//		}
//	}
//	err = c.store.SaveLeagues(ctx, leagues)
//	if err != nil {
//		c.log.Error(err)
//	}
//	err = c.store.SaveEvents(ctx, events)
//	if err != nil {
//		c.log.Error(err)
//	}
//	//c.log.Infow("", "", resp.GetData())
//	return nil
//}
