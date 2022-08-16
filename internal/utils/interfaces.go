package types_convertation

type SignalStopper interface {
	Stop()
}

type SignalStopperWithErr interface {
	Stop() error
}

type SignalCloser interface {
	Close()
}

type SignalCloserWithErr interface {
	Close() error
}
