package util

import (
	"sync"
	"time"
)

// exit is not nil if isTick==false, then Process() can have its inner loop
// otherwise,
// exit is nil, then Process() should have no inner loop
type RunnerSink interface {
	Process(exit chan bool)
}

type Runner struct {
	sync.RWMutex
	sink   RunnerSink
	isTick bool
	exitCh chan bool
}

func (l *Runner) Start(sink RunnerSink, isTick bool) {
	l.Close()
	l.sink = sink
	l.isTick = isTick
	l.exitCh = make(chan bool)
	go l.run()
}

func (l *Runner) Close() {
	close(l.exitCh)
}

func (l *Runner) run() {
	if !l.isTick {
		if l.sink != nil {
			l.sink.Process(l.exitCh)
		}
		return
	}

	tickChan := time.NewTicker(time.Millisecond * 10).C
	for quit := false; !quit; {
		select {
		case <-tickChan:
			if l.sink != nil {
				l.sink.Process(nil)
			}
		case <-l.exitCh:
			quit = true
		}
	}
}
