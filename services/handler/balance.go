package handler

import (
	"context"
	"time"
)

func (h *Handler) BalanceJob() {
	for {
		h.GetBalance()
		time.Sleep(time.Minute)
	}
}
func (h *Handler) GetBalance() float64 {
	got, b := h.balance.Get()
	if !b {
		go h.CheckBalance(false)
	}
	return got
}
func (h *Handler) CheckBalance(force bool) {
	//h.log.Info("got check balance run")
	h.balance.CheckBegin()
	defer h.balance.CheckDone()
	//h.log.Info("begin check balance")
	_, b := h.balance.Get()
	if b && !force {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	auth, err := h.auth.Auth(ctx)
	if err != nil {
		h.log.Info("get_auth_error")
		return
	}
	resp, err := h.client.Balance(auth, h.auth.GetAccount().Username)
	if err != nil {
		h.log.Error(err)
		return
	}
	//h.log.Infow("", "resp", resp.Data)
	for i := range resp.Data {
		if resp.Data[i].Key == "available_credit" {
			h.balance.Set(resp.Data[i].Value)
		} else if resp.Data[i].Key == "open_stakes" {
			h.balance.SetOutstanding(resp.Data[i].Value)
		}
	}
	//h.log.Infow("got_balance", "balance", h.balance.GetBalance(), "ou")
	h.log.Debugw("got_balance", "available", h.balance.GetBalance(), "outstanding", h.balance.GetOutstanding(),
		"full", h.balance.FullBalance(), "fill_factor", h.balance.CalcFillFactor())
}
