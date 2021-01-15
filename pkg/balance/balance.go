package balance

import (
	"github.com/aibotsoft/micro/util"
	"sync"
	"time"
)

const balanceMinPeriodSeconds = 60

type Balance struct {
	balance     float64
	outstanding float64
	last        time.Time
	mux         sync.RWMutex
	check       sync.Mutex
}

func (b *Balance) CheckBegin() {
	b.check.Lock()
}
func (b *Balance) CheckDone() {
	b.check.Unlock()
}

func (b *Balance) GetBalance() float64 {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.balance
}

func (b *Balance) GetOutstanding() float64 {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.outstanding
}

func (b *Balance) Get() (float64, bool) {
	b.mux.RLock()
	defer b.mux.RUnlock()
	var isFresh bool
	if time.Since(b.last).Seconds() < balanceMinPeriodSeconds {
		isFresh = true
	}
	return b.balance, isFresh
}

func (b *Balance) Set(bal float64) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.last = time.Now()
	b.balance = util.TruncateFloat(bal, 2)
}
func (b *Balance) SetOutstanding(value float64) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.outstanding = value
}

func (b *Balance) Sub(risk float64) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.balance = b.balance - risk
}
func (b *Balance) FullBalance() float64 {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return util.TruncateFloat(b.balance+b.outstanding, 2)
}
func (b *Balance) CalcFillFactor() float64 {
	return util.TruncateFloat(b.outstanding/b.FullBalance(), 2)
}
