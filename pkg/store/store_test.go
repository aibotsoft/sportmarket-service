package store

import (
	"context"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/logger"
	"github.com/aibotsoft/micro/sqlserver"
	"github.com/stretchr/testify/assert"
	"testing"
)

var s *Store

func TestMain(m *testing.M) {
	cfg := config.New()
	log := logger.New()
	db := sqlserver.MustConnectX(cfg)
	s = New(cfg, log, db)
	m.Run()
	s.Close()
}
func TestStore_GetEventById(t *testing.T) {
	got, err := s.GetEventById(context.Background(), "2020-05-30,821,30949")
	if assert.NoError(t, err) {
		assert.NotEmpty(t, got)
		t.Log(got)
	}
}

func TestStore_GetStartTime(t *testing.T) {
	got, err := s.GetStartTime(context.Background(), "2020-05-30,821,30949")
	if assert.NoError(t, err) {
		assert.NotEmpty(t, got)
		t.Log(got)
	}
}

func TestStore_GetResults(t *testing.T) {
	got, err := s.GetResults(context.Background())
	if assert.NoError(t, err) {
		assert.NotEmpty(t, got)
		t.Log(got)
	}
}
