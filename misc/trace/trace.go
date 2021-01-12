package trace

import (
	"fmt"
	"strings"
	"time"

	"github.com/fioncat/go-gendb/misc/log"
)

// Timer is used to record the execution time of each
// sub-process of a certain process and output it
// through the log.
type Timer struct {
	enable bool

	name  string
	start time.Time
	end   time.Time

	subs []*Timer
}

// NewTimer creates a new Timer
func NewTimer(op string) *Timer {
	if !log.IsEnable() {
		return &Timer{enable: false}
	}
	return &Timer{
		enable: true,
		name:   op,
		start:  time.Now(),
	}
}

// Start starts a new sub-process. If Start has been
// called before, it means that the last sub-process
// is over and the execution time of the previous process.
func (t *Timer) Start(op string) {
	if !t.enable {
		return
	}
	now := time.Now()
	if len(t.subs) > 0 {
		last := t.subs[len(t.subs)-1]
		last.end = now
	}
	sub := NewTimer(op)
	sub.start = now
	t.subs = append(t.subs, sub)
}

// Trace ends the whole process and shows the
// time-consuming of each sub-process.
func (t *Timer) Trace() {
	if !t.enable {
		return
	}
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
