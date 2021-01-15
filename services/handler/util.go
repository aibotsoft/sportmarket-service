package handler

import (
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/sportmarket-service/pkg/store"
	"github.com/pkg/errors"
	"regexp"
)

var betIdListRe = regexp.MustCompile(`betslip\/(\w+)\/(\w+)\/(.*?)\/(.*)`)

func ParseUrl(url string) (event store.UrlEvent, err error) {
	found := betIdListRe.FindStringSubmatch(url)
	if len(found) != 5 {
		return event, errors.Errorf("parse_url_error for: %q", url)
	}
	event.Sport = found[1]
	event.LeagueId = found[2]
	event.EventId = found[3]
	event.BetType = found[4]
	return event, nil
}

func (h *Handler) GetCurrency(sb *pb.Surebet) float64 {
	for _, currency := range sb.Currency {
		if currency.Code == h.auth.GetAccount().CurrencyCode {
			return currency.Value
		}
	}
	return 0
}
