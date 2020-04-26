/**
文件监听
*/
package sharebox

import (
	"time"

	"github.com/fsnotify/fsnotify"
)

type Batcher struct {
	*fsnotify.Watcher
	interval time.Duration
	done     chan struct{}

	Events chan []fsnotify.Event
}

func NewBatcher(interval time.Duration) (*Batcher, error) {
	watcher, err := fsnotify.NewWatcher()

	batcher := &Batcher{}
	batcher.Watcher = watcher
	batcher.interval = interval
	batcher.done = make(chan struct{}, 1)
	batcher.Events = make(chan []fsnotify.Event, 1)

	if err == nil {
		go batcher.run()
	}

	return batcher, err
}

func (b *Batcher) run() {
	tick := time.Tick(b.interval)
	evs := make([]fsnotify.Event, 0)
OuterLoop:
	for {
		select {
		case ev := <-b.Watcher.Events:
			//清除重复event
			var ok bool
			for _, v := range evs {
				if ev == v {
					ok = true
				}
			}
			if !ok {
				evs = append(evs, ev)
			}
		case <-tick:
			if len(evs) == 0 {
				continue
			}
			b.Events <- evs
			evs = make([]fsnotify.Event, 0)
		case <-b.done:
			break OuterLoop
		}
	}
	close(b.done)
}

func (b *Batcher) Close() {
	b.done <- struct{}{}
	b.Watcher.Close()
}
