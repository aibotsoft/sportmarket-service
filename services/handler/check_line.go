package handler

import (
	"context"
	"fmt"
	pb "github.com/aibotsoft/gen/fortedpb"
	api "github.com/aibotsoft/gen/sportmarketapi"
	"github.com/aibotsoft/micro/status"
	"github.com/aibotsoft/micro/util"
	"github.com/aibotsoft/sportmarket-service/pkg/store"
	"github.com/pkg/errors"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var TicketMap sync.Map

const (
	gbpEurRate         = 1.12
	goodEventThreshold = 0.64
	startsTimeSpread   = time.Minute * 6
	LockTimeOut        = time.Second * 34
)

var betslipCollection sync.Map
var exchangeList = []string{"bf", "bdaq", "mbook"}

func (h *Handler) GetBetSlip(ctx context.Context, side *pb.SurebetSide, u store.UrlEvent) (api.BetSlipData, error) {
	h.log.Infow("get_bet_slip", "sport", u.Sport, "event", u.EventId, "betType", u.BetType)
	key := fmt.Sprintf("%v:%v:%v", u.Sport, u.EventId, u.BetType)
	_, b := h.store.Cache.Get("betslip_request_error")
	if b {
		side.Check.Status = status.ServiceBusy
		side.Check.StatusInfo = "betslip_full"
		return api.BetSlipData{}, errors.Errorf(side.Check.StatusInfo)
	}
	resp, err := h.client.BetSlip(ctx, u.Sport, u.EventId, u.BetType)
	if err != nil {
		side.Check.Status = status.ServiceBusy
		side.Check.StatusInfo = "betslip_request_error"
		h.log.Infow(side.Check.StatusInfo, "err", err, "urlEvent", u, "sport", side.SportName, "league", side.LeagueName,
			"home", side.Home, "away", side.Away, "market", side.MarketName, "fortedPrice", side.Price)
		h.store.Cache.SetWithTTL("betslip_request_error", true, 1, time.Second*5)
		return api.BetSlipData{}, errors.Errorf(side.Check.StatusInfo)
	}
	if resp.GetStatus() != "ok" {
		side.Check.Status = status.StatusNotFound
		side.Check.StatusInfo = "betslip_status_not_ok"
		h.log.Infow(side.Check.StatusInfo, "resp", resp)
		return api.BetSlipData{}, errors.Errorf(side.Check.StatusInfo)
	}
	data := resp.GetData()
	betslipCollection.Store(key, data)
	h.log.Infow("bet_slip_response", "slipId", data.GetBetslipId(), "bet_type", data.GetBetType(),
		"bookies_with_offers_len", len(data.GetBookiesWithOffers()))
	return data, nil
}

func normalizeName(name string) string {
	if strings.Index(name, "(W)") != -1 {
		return strings.ReplaceAll(name, "(W)", "Women")
	} else if strings.Index(name, "(Res)") != -1 {
		return strings.ReplaceAll(name, "(Res)", "Reserves")
	}
	return name
}
func (h *Handler) FindSportMarketEvent(ctx context.Context, sb *pb.Surebet) (event store.UrlEvent, err error) {
	side := sb.Members[0]
	event, err = ParseUrl(side.Url)
	if err != nil {
		return
	}
	event.Starts, err = h.store.GetStartTime(ctx, event.EventId)
	if err != nil {
		side.Check.Status = status.StatusError
		side.Check.StatusInfo = "get_start_time_error"
		h.log.Infof("get_start_time_error: %v, EventId:%v, home:%v, away:%v, service:%v, sport:%v, league:%v", err, event.EventId, side.Home, side.Away, side.ServiceName, side.SportName, side.LeagueName)
		return
	}
	key := fmt.Sprintf("SportMarketEvent:%v:%v:%v:%v", sb.FortedSport, sb.Starts, sb.FortedHome, sb.FortedAway)
	h.store.Cache.SetWithTTL(key, event.EventId, 1, time.Hour*12)
	h.log.Infow("save_key", "key", key, "eventId", event.EventId)
	//if sb.FortedSport == "Футбол" {
	//	key := fmt.Sprintf("SportMarketEvent:%v:%v:%v:%v", sb.FortedSport, sb.Starts, sb.FortedHome, sb.FortedAway)
	//	h.store.Cache.SetWithTTL(key, event.EventId, 1, time.Hour*12)
	//	h.log.Infow("save_key", "key", key, "eventId", event.EventId)
	//}
	return
}
func (h *Handler) FindCloneEvent(ctx context.Context, sb *pb.Surebet) (event store.UrlEvent, err error) {
	side := sb.Members[0]

	key := fmt.Sprintf("SportMarketEvent:%v:%v:%v:%v", sb.FortedSport, sb.Starts, sb.FortedHome, sb.FortedAway)
	got, b := h.store.Cache.Get(key)
	if b {
		eventId := got.(string)
		event, err = h.store.GetEventById(ctx, eventId)
		if err != nil {
			side.Check.Status = status.StatusError
			side.Check.StatusInfo = "not_found_clone_in_db"
			return event, errors.New(side.Check.StatusInfo)
		}
		event.HomeScore = util.Compare(event.Home, normalizeName(side.Home))
		event.AwayScore = util.Compare(event.Away, normalizeName(side.Away))
		event.Avg = (event.HomeScore + event.AwayScore) / 2
		h.log.Infow("found_clone_event_from_sportmarket", "key", key, "eventId", eventId, "service", side.ServiceName, "event", event)
		if sb.FortedSport != "Футбол" && event.Avg < goodEventThreshold {
			side.Check.Status = status.StatusNotFound
			side.Check.StatusInfo = "found_clone_but_teams_diff"
			return event, errors.New(side.Check.StatusInfo)
		}

	} else {
		h.log.Infow("not_found_key", "key", key)

		parsedStarts, err := time.Parse(util.ISOFormat, sb.Starts)
		if err != nil {
			h.log.Error(err)
			side.Check.Status = status.StatusError
			side.Check.StatusInfo = "parse_start_time_error"
			return event, errors.New(side.Check.StatusInfo)
		}
		sportName := sb.FortedSport
		if sportName == "Регби" {
			sportName = side.SportName
		}
		events, err := h.store.SelectEventsByStarts(ctx, parsedStarts, startsTimeSpread, sportMap[sportName])
		if err != nil {
			h.log.Error(err)
			side.Check.Status = status.StatusError
			side.Check.StatusInfo = "not_found_event_by_start_time"
			return event, errors.New(side.Check.StatusInfo)
		}
		var goodEventList []store.UrlEvent
		var avgList []float64
		for i := range events {
			events[i].HomeScore = util.Compare(events[i].Home, normalizeName(side.Home))
			events[i].AwayScore = util.Compare(events[i].Away, normalizeName(side.Away))
			events[i].Avg = (events[i].HomeScore + events[i].AwayScore) / 2
			if events[i].Avg > goodEventThreshold {
				goodEventList = append(goodEventList, events[i])
				avgList = append(avgList, events[i].Avg)
			}
		}
		switch len(goodEventList) {
		case 0:
			h.log.Infow("not_found_good_event", "events_count", len(events), "starts", parsedStarts, "home", side.Home, "away", side.Away,
				"service", side.ServiceName, "fs", sb.FortedSport, "fs", sportMap[sb.FortedSport], "ss", side.SportName)
			for i := range events {
				homeScore := util.Compare(events[i].Home, side.Home)
				awayScore := util.Compare(events[i].Away, side.Away)
				avg := (homeScore + awayScore) / 2
				h.log.Debugw("", "home", events[i].Home, "hs", homeScore, "away", events[i].Away, "as", awayScore, "avg", avg)
			}
			side.Check.Status = status.StatusError
			side.Check.StatusInfo = "not_found_good_event"
			return event, errors.New(side.Check.StatusInfo)
		case 1:
			event = goodEventList[0]
		default:
			h.log.Infow("too_many_good_events", "home", side.Home, "away", side.Away, "league", side.LeagueName, "start", sb.Starts)

			for i := range goodEventList {
				homeScore := util.Compare(goodEventList[i].Home, side.Home)
				awayScore := util.Compare(goodEventList[i].Away, side.Away)
				leagueScore := util.Compare(goodEventList[i].LeagueName, side.LeagueName)
				avg := (homeScore + awayScore) / 2
				h.log.Debugw("", "home", goodEventList[i].Home, "hs", homeScore, "away", goodEventList[i].Away, "as", awayScore, "avg", avg,
					"starts", goodEventList[i].Starts,
					"league", goodEventList[i].LeagueName,
					"sideLeague", side.LeagueName,
					"leagueScore", leagueScore,
				)
			}

			sort.Float64s(avgList)
			//https://github.com/golang/go/wiki/SliceTricks
			for i := len(avgList)/2 - 1; i >= 0; i-- {
				opp := len(avgList) - 1 - i
				avgList[i], avgList[opp] = avgList[opp], avgList[i]
			}
			h.log.Info("sorted_avg_list:", avgList)
			if avgList[0] > 0.85 && avgList[1] < 0.7 {
				for i := range goodEventList {
					if goodEventList[i].Avg == avgList[0] {
						event = goodEventList[i]
					}
				}
			} else {
				side.Check.Status = status.StatusNotFound
				side.Check.StatusInfo = "too_many_good_events"
				return event, errors.New(side.Check.StatusInfo)
			}
		}
		h.log.Infow("found_event", "event", event)
	}

	err = Convert(sb, &event)
	if err != nil {
		h.log.Info("convert_error:", err, " market_name:", side.MarketName)
		side.Check.Status = status.StatusError
		side.Check.StatusInfo = "convert_error"
		return event, errors.New(side.Check.StatusInfo)
	}
	return event, nil
}
func (h *Handler) FindEvent(ctx context.Context, sb *pb.Surebet) (event store.UrlEvent, err error) {
	side := sb.Members[0]
	key := fmt.Sprintf("find_event:%v:%v:%v:%v:%v", sb.Starts, side.Home, side.Away, side.MarketName, side.ServiceName)
	got, b := h.store.Cache.Get(key)
	if b {
		return got.(store.UrlEvent), nil
	}
	switch side.ServiceName {
	case "SportMarket":
		event, err = h.FindSportMarketEvent(ctx, sb)
	case "Black":
		event, err = h.FindSportMarketEvent(ctx, sb)
	default:
		event, err = h.FindCloneEvent(ctx, sb)
	}
	if err != nil {
		return
	}
	h.store.Cache.SetWithTTL(key, event, 1, time.Minute*5)
	return
}
func (h *Handler) GetLock(sb *pb.Surebet) bool {
	side := sb.Members[0]
	key := side.MarketName + side.Home + side.Away + side.SportName + sb.Starts
	for i := 0; i < 40; i++ {
		got, b := h.store.Cache.Get(key)
		if b && got.(int64) != sb.SurebetId {
			time.Sleep(time.Millisecond * 50)
		} else {
			return h.store.SetVerifyWithTTL(key, sb.SurebetId, LockTimeOut)
		}
	}
	return false
}
func (h *Handler) CheckLine(ctx context.Context, sb *pb.Surebet) error {
	h.GetBalance()
	side := sb.Members[0]
	side.Check.AccountLogin = h.auth.GetAccount().Username
	side.Check.Currency = h.GetCurrency(sb)

	ok := h.GetLock(sb)
	if !ok {
		side.Check.Status = status.ServiceBusy
		side.Check.StatusInfo = "service_already_check_this_market"
		h.log.Debugw(status.ServiceBusy, "my_surebetId", sb.SurebetId)
		return nil
	}

	event, err := h.FindEvent(ctx, sb)
	if err != nil {
		h.log.Info(err)
		return nil
	}
	side.EventId = event.EventId

	auth, err := h.auth.Auth(ctx)
	if err != nil {
		side.Check.Status = status.BadBettingStatus
		side.Check.StatusInfo = "auth_session_error"
		return nil
	}

	var data api.BetSlipData
	var isBetslipNew bool
	key := fmt.Sprintf("%v:%v:%v", event.Sport, event.EventId, event.BetType)
	load, ok := betslipCollection.Load(key)
	if !ok {
		isBetslipNew = true
		data, err = h.GetBetSlip(auth, side, event)
		if err != nil {
			return nil
		}
	} else {
		data = load.(api.BetSlipData)
		betslipExpiry := int64(data.GetExpiryTs()) - time.Now().Unix()
		//h.log.Info("GetExpiryTs", betslipExpiry)
		switch {
		case betslipExpiry < 1:
			h.log.Infow("betslip_expiry_soon", "u", event)
			data, err = h.GetBetSlip(auth, side, event)
			if err != nil {
				return nil
			}
			isBetslipNew = true
		case betslipExpiry < 8:
			resp, err := h.client.RefreshBetSlip(auth, data.GetBetslipId())
			if err != nil || resp.Status != "ok" {
				h.log.Info("RefreshBetSlip_not_ok", err)
			} else {
				data.ExpiryTs = float64(time.Now().Unix()) + 39
				betslipCollection.Store(key, data)
				//h.log.Infow("refreshed_bet_slip", "new_expiry_ts", data.ExpiryTs, "u", event)
			}
		}
		//h.log.Infow("bet_slip_from_cache", "bookies_with_offers_len", len(data.GetBookiesWithOffers()), "accounts_len", len(data.GetAccounts()))
	}
	if len(data.GetBookiesWithOffers()) == 0 {
		side.Check.Status = status.StatusNotFound
		side.Check.StatusInfo = "bookies_with_offers_zero"
		return nil
	}
	for i := 0; i < 20; i++ {
		_, isPlaced := h.store.Cache.Get(data.GetBetslipId())
		if isPlaced {
			h.log.Infow("placed_recently", "GetBetslipId", data.GetBetslipId(), "fid", sb.FortedSurebetId)
			time.Sleep(time.Millisecond * 50)
		}
	}

	var list []api.WebsocketResponse
	fortedPrice := side.Price
	if event.IsLay {
		fortedPrice = util.TruncateFloat(side.Price/(side.Price-1), 2)
	}

waitLoop:
	for i := 0; i < 100; i++ {
		got, ok := betslipMap.Load(data.GetBetslipId())
		if ok {
			var bookieList []string
			list = got.([]api.WebsocketResponse)
			for i := range list {
				if !util.StringInList(list[i].Bookie, bookieList) {
					bookieList = append(bookieList, list[i].Bookie)
				}
			}
			if len(bookieList) >= len(data.GetBookiesWithOffers()) {
				//h.log.Infow("break", "list_len", len(list), "bookies_with_offers", len(data.GetBookiesWithOffers()), "bookieList", bookieList)
				break
			} else {
				for i := range list {
					if list[i].Price >= fortedPrice-0.01 {
						//h.log.Debugw("price_good_enough", "p", list[i].Price, "fp", fortedPrice, "max", list[i].Max)
						break waitLoop
					}
				}
			}
		}
		if !isBetslipNew {
			//h.log.Debug("betslip_not_new_so_break")
			break
		}
		time.Sleep(time.Millisecond * 30)
	}
	ticketKey := fmt.Sprintf("ticket:%v:%v:%v", side.ServiceName, side.Check.Id, side.Num)
	for i := range list {
		if list[i].Price > side.Check.Price && list[i].Price*list[i].Max > minVolume {
			TicketMap.Store(ticketKey, list[i])
			side.Check.Price = list[i].Price
		}
	}
	got, b := TicketMap.Load(ticketKey)
	if !b {
		side.Check.Status = status.StatusNotFound
		side.Check.StatusInfo = "price_list_empty"
		h.log.Infow(side.Check.StatusInfo, "sport", side.SportName, "league", side.LeagueName, "home", side.Home,
			"away", side.Away, "market", side.MarketName, "price", side.Price, "u", event)
		return nil
	}
	side.LeagueId, _ = strconv.ParseInt(event.LeagueId, 10, 64)
	eventSplit := strings.Split(event.EventId, ",")
	if len(eventSplit) == 3 {
		side.HomeId, _ = strconv.ParseInt(eventSplit[1], 10, 64)
		side.AwayId, _ = strconv.ParseInt(eventSplit[2], 10, 64)
	}
	err = h.store.GetStat(side)
	if err != nil {
		h.log.Error(err)
		side.Check.Status = status.StatusError
		side.Check.StatusInfo = "get_stat_for_event_error"
		return nil
	}
	ticket := got.(api.WebsocketResponse)
	//side.Check.SubService = strings.Replace(ticket.Bookie, "pin88", "pin", -1)
	side.Check.SubService = ticket.Bookie

	side.Check.Status = status.StatusOk
	side.Check.StatusInfo = fmt.Sprintf("chosen:%v, service_count:%v", ticket.Bookie, len(list))
	side.Check.Price = util.TruncateFloat(ticket.Price, 3)
	side.Check.MinBet = util.ToUSD(ticket.Min, side.Check.Currency)
	side.Check.MaxBet = util.TruncateFloat(util.ToUSD(ticket.Max, side.Check.Currency), 1)
	side.Check.Balance = util.ToUSDInt(h.GetBalance(), side.Check.Currency)
	side.Check.FillFactor = h.balance.CalcFillFactor()
	//h.log.Infow("found_event", "event", event, "side.Starts", side.Starts)
	side.Starts = event.Starts.Format(time.RFC3339)

	if util.StringInList(side.Check.SubService, exchangeList) && side.CheckCalc == nil {
		side.Check.MaxBet = util.TruncateFloat(side.Check.MaxBet-side.Check.MaxBet*10/100, 1)
	}

	h.log.Infow("check_ok",
		"p", side.Check.Price,
		"fp", fortedPrice,
		"sub", side.Check.SubService,
		"min", util.TruncateFloat(side.Check.MinBet, 2),
		"max", side.Check.MaxBet,
		"g", fmt.Sprintf("%v-%v-%v:%v", side.SportName, side.Home, side.Away, side.MarketName),
	//"CheckCalc", side.CheckCalc,
	//"fid", sb.FortedSurebetId,
	)
	return nil
}
