package auth

import (
	"context"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"github.com/aibotsoft/micro/logger"
	"github.com/aibotsoft/micro/sqlserver"
	"github.com/aibotsoft/sportmarket-service/pkg/store"
	"github.com/stretchr/testify/assert"
	"testing"
)

var a *Auth

func TestMain(m *testing.M) {
	cfg := config.New()
	log := logger.New()
	db := sqlserver.MustConnectX(cfg)
	sto := store.New(cfg, log, db)
	conf := config_client.New(cfg, log)
	a = New(cfg, log, sto, conf)
	m.Run()
}
func TestAuth_CheckLogin(t *testing.T) {
	err := a.CheckLogin(context.Background())
	assert.NoError(t, err)
}

func TestAuth_login(t *testing.T) {
	a.login(context.Background())
}
