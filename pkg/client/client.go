package client

import (
	"context"
	api "github.com/aibotsoft/gen/sportmarketapi"
	"github.com/aibotsoft/micro/config"
	"go.uber.org/zap"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type Client struct {
	cfg        *config.Config
	log        *zap.SugaredLogger
	apiConfig  *api.Configuration
	jar        *cookiejar.Jar
	tr         *http.Transport
	httpClient *http.Client
	*api.APIClient
}

func New(cfg *config.Config, log *zap.SugaredLogger) *Client {
	clientConfig := api.NewConfiguration()
	clientConfig.AddDefaultHeader("Origin", "https://m.sportmarket.com")
	jar, _ := cookiejar.New(nil)
	tr := &http.Transport{TLSHandshakeTimeout: 0 * time.Second, IdleConnTimeout: 0 * time.Second}
	httpClient := &http.Client{Jar: jar, Transport: tr}
	clientConfig.HTTPClient = httpClient
	clientConfig.Debug = cfg.Service.Debug
	client := api.NewAPIClient(clientConfig)
	a := &Client{cfg: cfg, log: log, jar: jar, APIClient: client, apiConfig: clientConfig, tr: tr}
	return a
}
func (c *Client) BetSlip(ctx context.Context, sport string, eventId string, betType string) (api.BetSlipResponse, error) {
	resp, _, err := c.BetApi.BetSlip(ctx).BetSlipRequest(api.BetSlipRequest{
		Sport:            sport,
		EventId:          eventId,
		BetType:          betType,
		EquivalentBets:   true,
		MultipleAccounts: true,
	}).Execute()
	return resp, err
}
func (c *Client) GetEvent(ctx context.Context) (api.GetEventsResponse, error) {
	req := api.GetEventsRequest{
		IncludePrices: false,
		AllEvents:     false,
		AllHcaps:      api.PtrBool(false),
	}
	resp, _, err := c.MarketApi.GetEvents(ctx).GetEventsRequest(req).Execute()
	return resp, err
}
func (c *Client) Events(ctx context.Context) (api.GetEventsResponse, error) {
	resp, _, err := c.MarketApi.GetEvents(ctx).GetEventsRequest(api.GetEventsRequest{IncludePrices: false, AllEvents: true}).Execute()
	return resp, err
}
func (c *Client) CheckLogin(ctx context.Context, session string) (api.LoginResponse, error) {
	resp, _, err := c.UserApi.CheckLogin(ctx, session).Full(true).Execute()
	if err != nil {
		return api.LoginResponse{}, err
	}
	return resp, nil
}
func (c *Client) Login(ctx context.Context, username string, password string) (api.LoginResponse, error) {
	resp, _, err := c.UserApi.Login(ctx).LoginRequest(api.LoginRequest{Username: username, Password: password, Full: true, Lang: "en"}).Execute()
	if err != nil {
		return api.LoginResponse{}, err
	}
	return resp, nil
}
func (c *Client) Balance(ctx context.Context, username string) (api.BalanceResponse, error) {
	resp, _, err := c.UserApi.Balance(ctx, username).Execute()
	return resp, err
}
func (c *Client) BetList(ctx context.Context) (api.BetListResponse, error) {
	resp, _, err := c.BetApi.BetList(ctx).Execute()
	return resp, err
}
func (c *Client) SettledBetList(ctx context.Context) (api.BetListResponse, error) {
	resp, _, err := c.BetApi.BetList(ctx).Status("reconciled").PageSize(250).OrderBy("placement_time desc").Execute()
	return resp, err
}
func (c *Client) PlaceBet(ctx context.Context, betslipId string, price float64, stake float64, requestUuid string, duration int64, noPutOffer bool) (api.PlaceBetResponse, error) {
	resp, _, err := c.BetApi.PlaceBet(ctx).PlaceBetRequest(api.PlaceBetRequest{
		BetslipId:               betslipId,
		Price:                   price,
		Stake:                   []interface{}{"EUR", stake},
		IgnoreSystemMaintenance: false,
		NoPutOfferExchange:      noPutOffer,
		RequestUuid:             requestUuid,
		Duration:                duration,
		//Accounts:                nil,
	}).Execute()
	if err != nil {
		apiError, ok := err.(api.GenericOpenAPIError)
		if ok {
			model := apiError.Model()
			c.log.Infow("place_bet_error", "BadRequestError", model)
			return resp, err
		}
		return resp, err
	}
	return resp, nil
}
func (c *Client) GetBetById(ctx context.Context, orderId int64) (api.PlaceBetResponse, error) {
	resp, _, err := c.BetApi.BetById(ctx, orderId).Execute()
	//c.log.Infow("resp", "", resp)
	return resp, err
}
func (c *Client) RefreshBetSlip(ctx context.Context, betslipId string) (api.RefreshBetSlipResponse, error) {
	resp, _, err := c.BetApi.RefreshBetSlip(ctx, betslipId).Execute()
	//c.log.Infow("resp", "", resp)
	return resp, err
}
func (c *Client) GetBetLog(ctx context.Context, orderId int64) (api.BetLogResponse, error) {
	resp, _, err := c.BetApi.BetLog(ctx, orderId).Execute()
	return resp, err

}
