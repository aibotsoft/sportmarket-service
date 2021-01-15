package handler

import (
	"context"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/sportmarket-service/pkg/store"
	"time"
)

const betListPeriod = 5 * time.Minute

var Commissions = map[string]float64{"bf": 2.3, "bdaq": 1.2, "mbook": 1.8, "3et": 0.5, "isn": 0.25}

func (h *Handler) GetResults(ctx context.Context) ([]pb.BetResult, error) {
	results, err := h.store.GetResults(ctx)
	return results, err
}

func (h *Handler) BetListJob() {
	time.Sleep(time.Second * 20)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := h.BetListRound(ctx)
		cancel()
		if err != nil {
			h.log.Error(err)
			time.Sleep(time.Minute)
		} else {
			time.Sleep(betListPeriod)
		}
	}
}

func (h *Handler) BetListRound(ctx context.Context) error {
	list, err := h.BetList(ctx)
	if err != nil {
		return err
	}
	//h.log.Infow("", "", list)
	err = h.store.SaveBetList(ctx, list)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) BetList(ctx context.Context) ([]store.Bet, error) {
	var bets []store.Bet
	auth, err := h.auth.Auth(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := h.client.SettledBetList(auth)
	if err != nil {
		return nil, err
	}
	for i := range resp.Data {
		//var profitLoss *float64
		var profitLoss2 float64
		//h.log.Infow("", "", resp.Data[i])

		//if resp.Data[i].GetOrderId() == 92571787 {
		//	h.log.Infow("", "", resp.Data[i])
		//
		//}
		for _, bet := range resp.Data[i].GetBets() {
			var profit float64
			profitLossList := bet.GetProfitLoss()
			if len(profitLossList) == 2 {
				profit = profitLossList[1].(float64)
			}
			bookie := bet.GetBookie()
			if profit > 0 {
				profit = profit - profit*Commissions[bookie]/100
			}
			profitLoss2 += profit
			//h.log.Infow("", "bookie", bookie, "profit", profit, "id", resp.Data[i].GetOrderId(),"profit2", profit2)
		}

		var wantStake, stake float64
		wantStakeList := resp.Data[i].GetWantStake()
		if len(wantStakeList) == 2 {
			wantStake = wantStakeList[1].(float64)
		}
		stakeList := resp.Data[i].GetStake()
		if len(stakeList) == 2 {
			stake = stakeList[1].(float64)
		}
		//profitLossList := resp.Data[i].GetProfitLoss()
		//if len(profitLossList) == 2 {
		//pl := profitLossList[1].(float64)
		//profitLoss = &pl
		//h.log.Infow("pl", "id", resp.Data[i].GetOrderId(), "pl", pl, "pl2", profitLoss2)
		//}
		b := store.Bet{
			Id:                 resp.Data[i].GetOrderId(),
			OrderType:          resp.Data[i].GetOrderType(),
			BetType:            resp.Data[i].GetBetType(),
			BetTypeDescription: resp.Data[i].GetBetTypeDescription(),
			BetTypeTemplate:    resp.Data[i].GetBetTypeTemplate(),
			Sport:              resp.Data[i].GetSport(),
			Placer:             resp.Data[i].GetPlacer(),
			WantPrice:          resp.Data[i].GetWantPrice(),
			Price:              resp.Data[i].GetPrice(),
			WantStake:          wantStake,
			Stake:              stake,
			ProfitLoss:         &profitLoss2,
			CcyRate:            resp.Data[i].GetCcyRate(),
			PlacementTime:      resp.Data[i].GetPlacementTime(),
			ExpiryTime:         resp.Data[i].GetExpiryTime(),
			Closed:             resp.Data[i].GetClosed(),
			CloseReason:        resp.Data[i].GetCloseReason(),
			Status:             resp.Data[i].GetStatus(),
			TakeStartingPrice:  resp.Data[i].GetTakeStartingPrice(),
			KeepOpenIr:         resp.Data[i].GetKeepOpenIr(),
			EventId:            resp.Data[i].GetEventInfo().EventId,
			HomeId:             resp.Data[i].GetEventInfo().HomeId,
			HomeTeam:           resp.Data[i].GetEventInfo().HomeTeam,
			AwayId:             resp.Data[i].GetEventInfo().AwayId,
			AwayTeam:           resp.Data[i].GetEventInfo().AwayTeam,
			CompetitionId:      resp.Data[i].GetEventInfo().CompetitionId,
			CompetitionName:    resp.Data[i].GetEventInfo().CompetitionName,
			CompetitionCountry: resp.Data[i].GetEventInfo().CompetitionCountry,
			StartTime:          resp.Data[i].GetEventInfo().StartTime,
			Date:               resp.Data[i].GetEventInfo().Date,
		}
		bets = append(bets, b)
		//h.log.Infow("b", "", b)
	}
	return bets, nil
}
