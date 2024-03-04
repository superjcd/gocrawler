package worker

type Signal int8

const (
	DummySignal = 1 << iota
	ContinueWithRetrySignal
	ContinueWithoutRetrySignal
	BreakWithPanicSignal
	BreakWithoutPanicSignal
)

const (
	KeepGoing = iota
	ContinueLoop
	BreakLoop
)
