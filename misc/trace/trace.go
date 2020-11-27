package trace

import (
	"fmt"
	"strings"
	"time"

	"github.com/fioncat/go-gendb/misc/log"
)

type Timer struct {
	name  string
	start time.Time
	end   time.Time

	subs []*Timer
}

func NewTimer(op string) *Timer {
	return &Timer{
		name:  op,
		start: time.Now(),
	}
}

func (t *Timer) Start(op string) {
	now := time.Now()
	if len(t.subs) > 0 {
		last := t.subs[len(t.subs)-1]
		last.end = now
	}
	sub := NewTimer(op)
	sub.start = now
	t.subs = append(t.subs, sub)
}

func (t *Timer) Trace() {
	now := time.Now()
	if len(t.subs) > 0 {
		last := t.subs[len(t.subs)-1]
		last.end = now
	}
	subInfo := make([]string, len(t.subs))
	for i, sub := range t.subs {
		took := sub.end.Sub(sub.start)
		info := fmt.Sprintf("%s: %v", sub.name, took)
		subInfo[i] = info
	}

	total := time.Since(t.start)
	info := fmt.Sprintf("[%s %v] %s", t.name,
		total, strings.Join(subInfo, ", "))
	log.Info(info)
}
