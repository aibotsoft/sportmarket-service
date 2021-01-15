package collector

import (
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"github.com/aibotsoft/micro/logger"
	"github.com/aibotsoft/micro/sqlserver"
	"github.com/aibotsoft/sportmarket-service/pkg/store"
	"github.com/aibotsoft/sportmarket-service/services/auth"
	"testing"
)

var c *Collector

func TestMain(m *testing.M) {
	cfg := config.New()
	log := logger.New()
	db := sqlserver.MustConnectX(cfg)
	sto := store.New(cfg, log, db)
	conf := config_client.New(cfg, log)
	au := auth.New(cfg, log, sto, conf)
	c = New(cfg, log, sto, au)
	m.Run()
}

//func TestCollector_Events(t *testing.T) {
//	err := c.CollectRound(context.Background())
//	assert.NoError(t, err)
//	//err = c.CollectRound()
//	//assert.NoError(t, err)
//}
