package handler

import (
	"context"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"github.com/aibotsoft/sportmarket-service/pkg/balance"
	"github.com/aibotsoft/sportmarket-service/pkg/client"
	"github.com/aibotsoft/sportmarket-service/pkg/store"
	"github.com/aibotsoft/sportmarket-service/services/auth"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Handler struct {
	cfg     *config.Config
	log     *zap.SugaredLogger
	client  *client.Client
	store   *store.Store
	auth    *auth.Auth
	Conf    *config_client.ConfClient
	balance balance.Balance
	wsConn  *websocket.Conn
}

func New(cfg *config.Config, log *zap.SugaredLogger, store *store.Store, auth *auth.Auth, conf *config_client.ConfClient) *Handler {
	h := &Handler{cfg: cfg, log: log, client: client.New(cfg, log), store: store, auth: auth, Conf: conf}
	return h
}
func (h *Handler) ReleaseCheck(ctx context.Context, sb *pb.Surebet) {
	side := sb.Members[0]
	key := side.MarketName + side.Home + side.Away + side.SportName + sb.Starts

	got, b := h.store.Cache.Get(key)
	if b && got.(int64) == sb.SurebetId {
		h.store.Cache.Del(key)
	}
}
func (h *Handler) Close() {
	h.store.Close()
	h.Conf.Close()
	if h.wsConn != nil {
		err := h.wsConn.Close()
		if err != nil {
			h.log.Error(err)
		}
	}
}
