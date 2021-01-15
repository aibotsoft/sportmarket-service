package handler

import (
	"encoding/json"
	"fmt"
	api "github.com/aibotsoft/gen/sportmarketapi"
	"github.com/aibotsoft/micro/util"
	"github.com/aibotsoft/sportmarket-service/pkg/store"
	"github.com/gorilla/websocket"
	"strconv"
	"strings"
	"sync"
	"time"
)

//const wsUrl = "wss://m.sportmarket.com/v1/stream/?token="
const wsUrl = "wss://m.sportmarket.com/cpricefeed/?token="
const minVolume = 28
const goodVolume = 200

var sportList = []string{"fb", "ih", "af", "basket", "rl", "ru", "baseball"}

var betslipMap sync.Map
var xRate sync.Map

type WsMsg struct {
	Ts   float64         `json:"ts"`
	Data [][]interface{} `json:"data"`
}

func (h *Handler) WebsocketConnect() error {
	var url = fmt.Sprintf("wss://m.sportmarket.com/cpricefeed/?token=%v&lang=en&prices_bookies=sbo,isn,sing2,pin88", h.auth.Session())

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	h.wsConn = conn
	return nil
}
func (h *Handler) WebsocketJob() {
	for {
		err := h.WebsocketConnect()
		if err != nil {
			h.log.Error(err)
			time.Sleep(time.Second * 20)
			continue
		}

		err = h.ReadLoop()
		if err != nil {
			h.log.Info(err)
			time.Sleep(time.Second * 10)
		}
	}
}

func (h *Handler) processEvent(data interface{}) {
	//h.log.Infow("data", "data", data)
	d := data.([]interface{})
	ss := d[1].([]interface{})
	sport := ss[0].(string)
	if !util.StringInList(sport, sportList) {
		//h.log.Info("fuck_sport: ", sport)
		return
	}
	var l store.League
	var home, away store.Team
	eventIdStr := ss[1].(string)
	eventSplit := strings.Split(eventIdStr, ",")
	if len(eventSplit) != 3 {
		h.log.Info("split_event_error: ", eventIdStr)
		return
	}
	home.Id, _ = strconv.ParseInt(eventSplit[1], 10, 64)
	away.Id, _ = strconv.ParseInt(eventSplit[2], 10, 64)

	em, ok := d[2].(map[string]interface{})
	if !ok {
		return
	}
	starts, ok := em["start_ts"].(string)
	if !ok {
		//h.log.Infow("", "start_ts_not_ok", em)
		return
	}
	st, err := time.Parse(time.RFC3339, starts)
	if err != nil {
		h.log.Info("time_error", st)
		return
	}
	home.Name = em["home"].(string)
	if home.Name == "" {
		return
	}
	away.Name = em["away"].(string)
	if away.Name == "" {
		return
	}
	l.Name = em["competition_name"].(string)
	if l.Name == "" {
		return
	}
	l.Id = int64(em["competition_id"].(float64))
	if l.Id == 0 {
		return
	}
	l.Country = em["country"].(string)
	l.Sport = sport

	var e = store.Event{
		Id:       eventIdStr,
		LeagueId: l.Id,
		HomeId:   home.Id,
		AwayId:   away.Id,
		Starts:   st,
	}
	//h.log.Infow("", "asdf", sport, "", eventIdStr, "", em, "home", home, "away", away, "starts", starts, "e", e, "l", l)
	//h.log.Infow("", "e", e)
	h.store.SaveLeague(l)
	h.store.SaveTeam(home)
	h.store.SaveTeam(away)
	h.store.SaveEvent(e)
}
func (h *Handler) ReadLoop() error {
	for {
		_, b := h.store.Cache.Get("ts")
		if !b {
			ts := util.UnixMsNow()
			h.store.Cache.SetWithTTL("ts", ts, 1, time.Second*5)
			pingMsg := fmt.Sprintf(`["echo",%v]`, ts)
			//h.log.Info("need_ping", pingMsg)
			err := h.wsConn.WriteMessage(1, []byte(pingMsg))
			if err != nil {
				h.log.Info("WriteMessage_error: ", err)
			}
		}
		_, message, err := h.wsConn.ReadMessage()
		if err != nil {
			return err
		}
		var messages [][]interface{}
		err = json.Unmarshal(message, &messages)
		if err != nil {
			h.log.Info(err)
			continue
		}
		for _, msg := range messages {
			if msg[0] == "event" {
				h.processEvent(msg)
			} else if msg[0] == "api" {
				apiMap, ok := msg[1].(map[string]interface{})
				if !ok {
					continue
				}
				dataTypeSlice, ok := apiMap["data"].([]interface{})
				if !ok {
					continue
				}
				for _, data := range dataTypeSlice {
					d := data.([]interface{})
					if d[0] == "pmm" {
						//h.log.Infow("", "api", d[1])
						wsm := h.processMsg(d[1])
						//h.log.Infow("", "", wsm)
						got, ok := betslipMap.Load(wsm.GetBetslipId())
						if !ok {
							if wsm.Price > 1 {
								betslipMap.Store(wsm.GetBetslipId(), []api.WebsocketResponse{wsm})
							} else {
								//h.log.Infow("wsm_price_missing", "wsm", wsm)
							}
						} else {
							var found bool
							list := got.([]api.WebsocketResponse)
							for i := range list {
								if list[i].Bookie == wsm.Bookie && list[i].BetType == wsm.BetType {
									list[i] = wsm
									found = true
								}
							}
							if !found {
								if wsm.Price > 1 {
									list = append(list, wsm)
								} else {
									//h.log.Infow("wsm_price_missing", "wsm", wsm)
								}
							}
							betslipMap.Store(wsm.GetBetslipId(), list)
						}
					}
				}
			} else if msg[0] == "offers_hcap" {
				h.log.Infow("hcap", "msg", msg)
			}
		}

	}
}
func calcWeightedPrice(prices []float64, volumes []float64) float64 {
	var vol, pSum float64
	for i := 0; i < len(prices); i++ {
		pSum += prices[i] * volumes[i]
		vol += volumes[i]
	}
	return pSum / vol
}

func (h *Handler) processMsg(data interface{}) (price api.WebsocketResponse) {
	respMap, ok := data.(map[string]interface{})
	if !ok {
		return
	}
	priceList := respMap["price_list"].([]interface{})
	if len(priceList) == 0 {
		return
	}
	var min, max, bestPrice float64
	var prices, volumes []float64
	for i := range priceList {
		if i >= 3 {
			break
		}
		priceMap := priceList[i].(map[string]interface{})
		bookie := priceMap["bookie"].(map[string]interface{})
		effective := priceMap["effective"].(map[string]interface{})

		price := effective["price"].(float64)
		if i > 0 && bestPrice > 0 {
			curr := price * 100 / bestPrice
			if curr < 98 {
				break
			}
		}
		minBookieSlice := bookie["min"].([]interface{})
		maxBookieSlice := bookie["max"].([]interface{})
		maxBookie := maxBookieSlice[1].(float64)
		min = minBookieSlice[1].(float64)
		if maxBookie < min*2 && i > 0 {
			continue
		}
		max += maxBookie
		prices = append(prices, price)
		volumes = append(volumes, maxBookie)
		if i == 0 {
			bestPrice = price
		}
		if price*maxBookie >= goodVolume {
			break
		}
	}
	weightedPrice := calcWeightedPrice(prices, volumes)
	price.Price = weightedPrice
	price.Max = max
	price.Min = min
	price.BetslipId = respMap["betslip_id"].(string)
	price.Username = respMap["username"].(string)
	price.EventId = respMap["event_id"].(string)
	price.Sport = respMap["sport"].(string)
	price.BetType = respMap["bet_type"].(string)
	price.Bookie = respMap["bookie"].(string)
	price.Status = respMap["status"].(map[string]interface{})["code"].(string)
	//h.log.Infow("", "min", min,"max", max, "name", price.Bookie, "weightedPrice", weightedPrice)
	//if len(prices) > 1 {
	//	h.log.Infow("", "prices", prices, "volumes", volumes, "min", min,"max", max, "name", price.Bookie, "weightedPrice", weightedPrice)
	//}
	return price
}
