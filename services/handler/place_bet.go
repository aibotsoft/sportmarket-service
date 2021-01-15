package handler

import (
	"context"
	"fmt"
	pb "github.com/aibotsoft/gen/fortedpb"
	api "github.com/aibotsoft/gen/sportmarketapi"
	"github.com/aibotsoft/micro/status"
	"github.com/aibotsoft/micro/util"
	"github.com/google/uuid"
	"strconv"
	"time"
)

const PriceReduction = 6

func (h *Handler) PlaceBet(ctx context.Context, sb *pb.Surebet) error {
	side := sb.Members[0]
	auth, err := h.auth.Auth(ctx)
	if err != nil {
		side.Bet.Status = status.BadBettingStatus
		side.Bet.StatusInfo = "auth_session_error"
		return nil
	}
	err = h.store.SaveCheck(sb)
	if err != nil {
		h.log.Info("save_check_error: ", err)
		side.Bet.Status = status.StatusError
		side.Bet.StatusInfo = "save_check_error"
		return nil
	}
	ticketKey := fmt.Sprintf("ticket:%v:%v:%v", side.ServiceName, side.Check.Id, side.Num)

	got, ok := TicketMap.Load(ticketKey)
	if !ok {
		side.Bet.Status = status.StatusError
		side.Bet.StatusInfo = "check_ticket_not_found"
		return nil
	}
	ticket := got.(api.WebsocketResponse)
	//price := util.TruncateFloat(side.Check.Price, 3)
	//priceCommission := (side.Check.Price - 1.0) * Commissions[side.Check.SubService] / 100
	price := util.TruncateFloat(side.Check.Price-side.Check.Price*0.4/100, 2)

	var placeTimeout int64 = 4
	if !side.CheckCalc.IsFirst {
		price = util.TruncateFloat(price-price*PriceReduction/100, 3)
		placeTimeout = 30
	}

	stake := util.AdaptStake(side.CheckCalc.Stake, side.Check.Currency, side.BetConfig.RoundValue)

	requestId := uuid.New().String()
	h.log.Infow("begin_place_bet", "price", price, "stake", stake, "ticket", ticket, "ticket.Min", ticket.Min)
	if stake <= ticket.Min {
		h.log.Infow("stake_lower_min_bet", "price", price, "stake", stake, "ticket", ticket, "ticket.Min", ticket.Min)
		stake = util.TruncateFloat(ticket.Min+0.1, 2)
	}

	resp, err := h.client.PlaceBet(auth, ticket.BetslipId, price, stake, requestId, placeTimeout, side.CheckCalc.IsFirst)
	if err != nil {
		side.Bet.Status = status.StatusError
		side.Bet.StatusInfo = "place_bet_request_error"
		h.log.Info(err)
		return nil
	}
	if resp.GetStatus() != "ok" {
		side.Bet.Status = status.StatusError
		side.Bet.StatusInfo = "place_bet_response_status_not_ok"
		h.log.Infow("place_bet_response_status_not_ok", "resp", resp)
		return nil
	}
	betApiId := resp.Data.GetOrderId()
	h.log.Infow("PlaceBet_response", "data", resp.Data, "betApiId", betApiId)

	var order api.OrderItem

	for i := 0; i < 10*int(placeTimeout); i++ {
		time.Sleep(time.Millisecond * 150)
		var placedStake, unplaced, inprogress float64
		resp, err := h.client.GetBetById(auth, betApiId)
		if err != nil {
			h.log.Infow("GetBetById_error", "err", err)
			continue
		}
		order = resp.Data
		if order.GetClosed() == true && order.GetStatus() != "pending" {
			break
		}
		successList := order.GetBetBarValues().Success
		unplacedList := order.GetBetBarValues().Unplaced
		inprogressList := order.GetBetBarValues().Inprogress
		if len(successList) == 2 {
			placedStake = successList[1].(float64)
		}
		if len(unplacedList) == 2 {
			unplaced = unplacedList[1].(float64)
		}
		if len(inprogressList) == 2 {
			inprogress = inprogressList[1].(float64)
		}
		if placedStake >= (stake - 2) {
			h.log.Debugw("placed_stake_near", "stake", stake, "placed", placedStake)
			break
		}

		//if placedStake> 0 && inprogress == 0 && unplaced <= minBet {
		//	h.log.Infow("unplaced_too_low_should_break", "placed", placedStake, "unplaced", unplaced, "inprogress", inprogress, "min_bet", minBet)
		//}

		//log, err := h.client.GetBetLog(auth, betApiId)
		//if err != nil || log.Status != "ok" {
		//	h.log.Info("get_bet_log_error: ", err)
		//} else {
		//	//h.log.Infow("log_response", "log_data", log.Data)
		//}

		h.log.Infow("GetBetById_response",
			"order_id", order.GetOrderId(),
			"want_price", order.GetWantPrice(),
			"want_stake", order.GetWantStake(),
			"status", order.GetStatus(),
			"bet_type", order.GetBetType(),
			"bet_type_description", order.GetBetTypeDescription(),
			"placed", placedStake, "unplaced", unplaced, "inprogress", inprogress)
		time.Sleep(time.Millisecond * 100)
	}
	side.Bet.ApiBetId = strconv.FormatInt(betApiId, 10)

	if order.GetClosed() == true {
		if len(order.GetStake()) == 2 {
			placedStake := order.GetStake()[1].(float64)
			if placedStake > 0 {
				side.Bet.Status = status.StatusOk
				side.Bet.StatusInfo = fmt.Sprintf("status:%v, close_reason:%v", order.GetStatus(), order.GetCloseReason())
				side.Bet.Stake = util.ToUSD(placedStake, side.Check.Currency)
				if order.GetPrice() > 0 {
					side.Bet.Price = order.GetPrice()
				} else if order.GetWantPrice() > 0 {
					h.log.Warnw("price_is_null", "Price", order.GetPrice())
					side.Bet.Price = order.GetWantPrice()
				} else {
					h.log.Warnw("price_and_want_price_is_null", "price", order.GetPrice(), "WantPrice", order.GetWantPrice())
					side.Bet.Price = price
				}
			}
		} else {
			side.Bet.Status = status.StatusNotAccepted
			side.Bet.StatusInfo = fmt.Sprintf("status:%v, close_reason:%v", order.GetStatus(), order.GetCloseReason())
		}
	} else {
		successList := order.GetBetBarValues().Success
		if len(successList) == 2 {
			placedStake := successList[1].(float64)
			if placedStake > 0 {
				side.Bet.Status = status.StatusOk
				side.Bet.StatusInfo = fmt.Sprintf("status:%v, close_reason:%v", order.GetStatus(), order.GetCloseReason())
				side.Bet.Stake = util.ToUSD(placedStake, side.Check.Currency)
				if order.GetPrice() > 0 {
					side.Bet.Price = order.GetPrice()
				} else if order.GetWantPrice() > 0 {
					h.log.Warnw("price_is_null", "Price", order.GetPrice())
					side.Bet.Price = order.GetWantPrice()
				} else {
					h.log.Warnw("price_and_want_price_is_null", "price", order.GetPrice(), "WantPrice", order.GetWantPrice())
					side.Bet.Price = price
				}
			} else {
				side.Bet.Status = status.StatusNotAccepted
				side.Bet.StatusInfo = fmt.Sprintf("status:%v, close_reason:%v", order.GetStatus(), order.GetCloseReason())
			}
		}
	}
	betslipMap.Delete(ticket.BetslipId)
	//save BetslipId of placed bet to pause repeated try
	h.store.Cache.SetWithTTL(ticket.BetslipId, true, 1, time.Millisecond*800)

	side.Bet.Done = util.UnixMsNow()
	h.log.Infow("side_bet", "bet", side.Bet, "fid", sb.FortedSurebetId, "betslip", ticket.BetslipId, "time", side.Bet.Done-side.ToBet.Id)
	err = h.store.SaveBet(sb)
	if err != nil {
		h.log.Error(err)
	}
	return nil
}
