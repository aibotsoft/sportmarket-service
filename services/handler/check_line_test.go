package handler

import (
	"context"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"github.com/aibotsoft/micro/logger"
	"github.com/aibotsoft/micro/sqlserver"
	"github.com/aibotsoft/micro/status"
	"github.com/aibotsoft/micro/util"
	"github.com/aibotsoft/sportmarket-service/pkg/store"
	"github.com/aibotsoft/sportmarket-service/services/auth"
	"github.com/stretchr/testify/assert"
	"testing"
)

var h *Handler

func TestMain(m *testing.M) {
	cfg := config.New()
	log := logger.New()
	db := sqlserver.MustConnectX(cfg)
	sto := store.New(cfg, log, db)
	conf := config_client.New(cfg, log)
	au := auth.New(cfg, log, sto, conf)
	h = New(cfg, log, sto, au, conf)
	go h.WebsocketJob()
	m.Run()
	h.Close()
}

func sbHelper(t *testing.T) *pb.Surebet {
	t.Helper()
	return &pb.Surebet{
		Starts:      "2020-06-25 22:05:00+03:00",
		FortedSport: "Футбол",
		Currency:    []pb.Currency{{Code: "USD", Value: 1}, {Code: "EUR", Value: 0.93}},
		Members: []*pb.SurebetSide{{
			ServiceName: "SportMarket",
			//Origin:      "Betdaq",
			//SportName:   "Football: Full-Time",
			//LeagueName:  "Germany Bundesliga",
			//Home:        "Nk Varazdin",
			//Away:        "Inter Zapresic",
			//MarketName:  "2",
			Price: 5.4,
			//Url:         "https://www.betdaq.com/exchange/soccer/7301158",
			Url:       "https://m.sportmarket.com/betslip/fb/179/2020-08-10,231,202/for,odd",
			Check:     &pb.Check{Id: util.UnixMsNow()},
			Bet:       &pb.Bet{},
			ToBet:     &pb.ToBet{Id: util.UnixMsNow(), TryCount: 0},
			BetConfig: &pb.BetConfig{RoundValue: 0.01},
			CheckCalc: &pb.CheckCalc{
				Status:   "Ok",
				MaxStake: 1,
				MinStake: 1,
				MaxWin:   2,
				Stake:    0.7,
				Win:      1,
				IsFirst:  true,
			},
		}},
	}
}

func TestHandler_CheckLine(t *testing.T) {
	sb := sbHelper(t)
	err := h.CheckLine(context.Background(), sb)
	if assert.NoError(t, err) {
		side := sb.Members[0]
		if side.Check.Status != status.StatusOk {
			t.Log(side.Check)
		} else {
			t.Log(side.Check, side.HomeId, side.AwayId, side.Starts, side.EventId)
			t.Log(util.UnixMsNow() - side.Check.Id)
		}
	}
}

func TestHandler_CheckLine_Betfair(t *testing.T) {
	e, _ := h.store.GetEventById(context.Background(), "2020-09-12,812,890")
	t.Log(e)
	sb := &pb.Surebet{
		Starts:      e.Starts.Format(util.ISOFormat),
		FortedSport: "Футбол",
		Currency:    []pb.Currency{{Code: "USD", Value: 1}, {Code: "EUR", Value: 0.93}},
		Members: []*pb.SurebetSide{{
			ServiceName: "Betfair",
			SportName:   "Football: Full-Time",
			LeagueName:  e.LeagueName,
			Home:        e.Home,
			Away:        e.Away,
			MarketName:  "0:0",
			Price:       9,
			//Url:       "https://m.sportmarket.com/betslip/fb/179/2020-08-10,231,202/for,odd",
			Check:     &pb.Check{Id: util.UnixMsNow()},
			ToBet:     &pb.ToBet{Id: util.UnixMsNow(), TryCount: 0},
			BetConfig: &pb.BetConfig{RoundValue: 0.01},
		}},
	}
	err := h.CheckLine(context.Background(), sb)
	if assert.NoError(t, err) {
		//side := sb.Members[0]
		//if side.Check.Status != status.StatusOk {
		//	t.Log(side.Check)
		//} else {
		//	t.Log(side.Check, side.HomeId, side.AwayId, side.Starts, side.EventId)
		//	t.Log(util.UnixMsNow() - side.Check.Id)
		//}
	}
}

//func TestHandler_WebsocketClient(t *testing.T) {
//	h.WebsocketClient()
//}
