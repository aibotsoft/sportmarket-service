package client

import (
	"context"
	api "github.com/aibotsoft/gen/sportmarketapi"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var c *Client
var session string
var auth context.Context

func TestMain(m *testing.M) {
	cfg := config.New()
	log := logger.New()
	session = os.Getenv("test_session")
	ctx := context.Background()
	auth = context.WithValue(ctx, api.ContextAPIKeys, map[string]api.APIKey{"session": {Key: session}})
	c = New(cfg, log)
	m.Run()
}

func TestClient_CheckLogin(t *testing.T) {
	_, err := c.CheckLogin(auth, session)
	assert.NoError(t, err)
}

func TestClient_BetSlip(t *testing.T) {
	got, err := c.BetSlip(auth, "fb", "2020-05-25,26962,873", "for,h")
	if assert.NoError(t, err) {
		assert.NotEmpty(t, got)
		t.Log(got)
	}
}

func TestClient_BetList(t *testing.T) {
	got, err := c.BetList(auth)
	if assert.NoError(t, err) {
		assert.NotEmpty(t, got)
		t.Log(got.Data)
	}
}

func TestClient_PlaceBet(t *testing.T) {
	uniqueRequestId := uuid.New().String()
	got, err := c.PlaceBet(auth, "8960e9ac657d4ab9ab21b83e87b5423d", 1.5, 3, uniqueRequestId, 3)
	t.Log("", got, err)

	if assert.NoError(t, err) {
		assert.NotEmpty(t, got)
	}
}

func TestClient_GetBetById(t *testing.T) {
	got, err := c.GetBetById(auth, 90227958)
	if assert.NoError(t, err) {
		assert.NotEmpty(t, got)
		//t.Log(got)
	}
}

func TestClient_RefreshBetSlip(t *testing.T) {
	got, err := c.RefreshBetSlip(auth, "819063a4d30042b2a3dd938036e37f2b")
	if assert.NoError(t, err) {
		assert.NotEmpty(t, got)
		t.Log(got)
	}
}

func TestClient_GetEvent(t *testing.T) {
	got, err := c.GetEvent(auth)
	if assert.NoError(t, err) {
		assert.NotEmpty(t, got)
		t.Log(got.GetData())
	}
}
