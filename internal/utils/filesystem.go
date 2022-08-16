package types_convertation

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"

	"github.com/rs/zerolog"
)

func RootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	return filepath.Dir(filepath.Dir(d))
}

func closeAll(l ...interface{}) {
	for _, cl := range l {
		switch closer := cl.(type) {
		case SignalStopper:
			closer.Stop()
		case SignalCloser:
			closer.Close()
		case SignalCloserWithErr:
			err := closer.Close()
			if err != nil {
				switch logger := l[len(l)-1].(type) {
				case *zerolog.Logger:
					logger.Error().Msgf("🔥 close SignalCloserWithErr object type: %T, error: %v", closer, err)
				case *log.Logger:
					logger.Printf("🔥 close SignalCloserWithErr object type: %T, error: %v", closer, err)
				}
			}
		case SignalStopperWithErr:
			err := closer.Stop()
			if err != nil {
				switch logger := l[len(l)-1].(type) {
				case *zerolog.Logger:
					logger.Error().Msgf("🔥 stop SignalStopperWithErr object type: %T, error: %v", closer, err)
				case *log.Logger:
					logger.Printf("🔥 stop SignalStopperWithErr object type: %T, error: %v", closer, err)
				}
			}
		}
	}
}

var doOnce sync.Once

func OsSignalHandler(cancel context.CancelFunc, callbacks []func(), l ...interface{}) {
	doOnce.Do(func() {
		var Stop = make(chan os.Signal, 1)

		signal.Notify(Stop,
			syscall.SIGTERM,
			syscall.SIGINT,
			syscall.SIGABRT,
		)

		for range Stop {
			if cancel != nil {
				cancel()
			}

			closeAll(l...)

			for _, callback := range callbacks {
				callback()
			}

			return
		}
	})
}
