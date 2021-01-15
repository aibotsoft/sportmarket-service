package handler

import (
	"context"
	"github.com/aibotsoft/micro/status"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHandler_PlaceBet(t *testing.T) {
	ctx := context.Background()
	sb := sbHelper(t)
	err := h.CheckLine(ctx, sb)
	if assert.NoError(t, err) {
		side := sb.Members[0]
		if side.Check.Status != status.StatusOk {
			t.Log(side)
			return
		}
		//side.Check.Price = side.Check.Price + 0.1
		err := h.PlaceBet(ctx, sb)
		if assert.NoError(t, err) {
			if side.Bet.Status != status.StatusOk {
				t.Log(side.Bet)
			}
		}
	}
}
