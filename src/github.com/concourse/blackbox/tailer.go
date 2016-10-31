package blackbox

import (
	"log"
	"os"
	"time"

	"github.com/hpcloud/tail"
	"github.com/hpcloud/tail/watch"

	"github.com/concourse/blackbox/syslog"
)

type Tailer struct {
	Path    string
	Tag     string
	Drainer syslog.Drainer
}

func (tailer *Tailer) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	watch.POLL_DURATION = 1 * time.Second

	t, err := tail.TailFile(tailer.Path, tail.Config{
		Follow: true,
		ReOpen: true,
		Poll:   true,
		Location: &tail.SeekInfo{
			Offset: 0,
			Whence: os.SEEK_END,
		},
	})

	if err != nil {
		return err
	}
	defer t.Cleanup()

	close(ready)

	for {
		select {
		case line, ok := <-t.Lines:
			if !ok {
				log.Println("lines flushed; exiting tailer")
				return nil
			}

			tailer.Drainer.Drain(line.Text, tailer.Tag)
		case <-signals:
			return t.Stop()
		}
	}

	return nil
}
