package auth

import (
	"context"
	pb "github.com/aibotsoft/gen/confpb"
	api "github.com/aibotsoft/gen/sportmarketapi"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"github.com/aibotsoft/sportmarket-service/pkg/client"
	"github.com/aibotsoft/sportmarket-service/pkg/store"
	"github.com/aibotsoft/sportmarket-service/pkg/token"
	"go.uber.org/zap"
	"log"
	"sync"
)

type Auth struct {
	cfg           *config.Config
	log           *zap.SugaredLogger
	store         *store.Store
	client        *client.Client
	conf          *config_client.ConfClient
	account       pb.Account
	token         token.Token
	bettingStatus bool
	tokenLock     sync.Mutex
}

func New(cfg *config.Config, log *zap.SugaredLogger, store *store.Store, conf *config_client.ConfClient) *Auth {
	return &Auth{cfg: cfg, log: log, store: store, client: client.New(cfg, log), conf: conf}
}
func (a *Auth) GetAccount() pb.Account {
	if a.account.Id == 0 {
		var err error
		a.account, err = a.conf.GetAccount(context.Background(), a.cfg.Service.Name)
		if err != nil {
			log.Panic(err)
		}
	}
	return a.account
}
func (a *Auth) Session() string {
	a.tokenLock.Lock()
	defer a.tokenLock.Unlock()
	if a.token.IsEmpty() {
		err := a.LoadToken(context.Background())
		if err != nil {
			a.log.Info(err)
			return ""
		}
	}
	return a.token.Session
}
func (a *Auth) LoadToken(ctx context.Context) (err error) {
	a.token, err = a.store.LoadToken(ctx)
	return
}
func (a *Auth) Auth(ctx context.Context) (context.Context, error) {
	if a.token.IsEmpty() {
		err := a.LoadToken(ctx)
		if err != nil {
			return nil, err
		}
	}
	auth := context.WithValue(ctx, api.ContextAPIKeys, map[string]api.APIKey{"session": {Key: a.token.Session}})
	return auth, nil
}

func (a *Auth) CheckLogin(ctx context.Context) (err error) {
	resp, err := a.client.CheckLogin(ctx, a.Session())
	a.log.Infow("", "", resp)
	return
}

func (a *Auth) LoginRound(ctx context.Context) {
	err := a.CheckLogin(ctx)
	if err != nil {
		a.log.Infow("check_login_error", "err", err)
		a.login(ctx)
	}

}
func (a *Auth) login(ctx context.Context) {
	a.log.Infow("begin_login", "account", a.account)
	resp, _, err := a.client.UserApi.Login(ctx).LoginRequest(api.LoginRequest{
		Username: a.account.Username,
		Password: a.account.Password,
		Full:     true,
		Lang:     "en",
	}).Execute()
	if err != nil {
		a.log.Error(err)
	}
	a.log.Infow("login_resp", "resp", resp)
	data := resp.GetData()
	session := data.GetSessionId()
	a.log.Infow("", "session", session)

}
