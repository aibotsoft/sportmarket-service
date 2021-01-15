package handler

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHandler_BetList(t *testing.T) {
	got, err := h.BetList(context.Background())
	if assert.NoError(t, err) {
		assert.NotEmpty(t, got)
		//t.Log(got)
	}
}

func TestHandler_BetListRound(t *testing.T) {
	err := h.BetListRound(context.Background())
	if assert.NoError(t, err) {
	}
}

func TestHandler_GetResults(t *testing.T) {
	got, err := h.GetResults(context.Background())
	if assert.NoError(t, err) {
		assert.NotEmpty(t, got)
		h.log.Infow("", "", got)
	}
}
