package handler

import (
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/sportmarket-service/pkg/store"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

//func TestConvert(t *testing.T) {
//	got, err := Convert("1")
//	if assert.NoError(t, err) {
//		t.Log(got)
//	}
//	//assert.Equal(t, -0.5, *got.Handicap)
//}
func TestConvert(t *testing.T) {
	split := strings.Split(testNames, "\n")
	for _, name := range split {
		if name != "" {
			//t.Log(name)
			var event store.UrlEvent
			sb := &pb.Surebet{
				FortedSport: "Футбол",
				Members: []*pb.SurebetSide{{
					MarketName: name,
				}},
			}
			err := Convert(sb, &event)
			assert.NoError(t, err)
			t.Log(name, ":", event.BetType, event.Handicap)
		}
	}
}

var testNames = `
1
2
Х
1Х
Х2
12
П1
П2
Ф1(-10)
Ф2(0,25)
ТБ(170)
ТМ(170)
ИТ1Б(2,5)
ИТ1М(2,5)
ИТ2Б(2,5)
ИТ2М(2,5)
Чёт
Нечёт
Ф1(0,5)
1(1:0)

Ф1(-1,5)
1(0:1)
Ф2(0,5)
2(0:1)
Ф2(-1,5)
2(1:0)
`
