package signal

import (
	"os"
	"os/signal"
)

type SignalHandlerFunc func()

var (
	InterruptSignalFunc = func() {
		os.Exit(0)
	}
)

type Signal struct {
	handlers map[string]SignalHandlerFunc
	sigCh    chan os.Signal
}

type SignalManager struct {
	sigs map[os.Signal]*Signal
}

func (s *SignalManager) SignalHandlerFuncWrapper(notifyCh <-chan os.Signal, f SignalHandlerFunc) SignalHandlerFunc {
	return func() {
		<-notifyCh
		f()
	}
}

func NewManager() *SignalManager {
	s := &SignalManager{
		sigs: make(map[os.Signal]*Signal),
	}
	return s
}

// 如果没有任何信号监听，直接返回
func (s *SignalManager) Wait() {
	if len(s.sigs) == 0 {
		println("[WARN] no signal watcher")
		return
	}
	for _, sig := range s.sigs {
		for _, sh := range sig.handlers {
			go sh()
		}
	}
	s.AddWatcher(os.Interrupt, "_internal_", InterruptSignalFunc)
	s.sigs[os.Interrupt].handlers["_internal_"]()
}

// SignalHandlerFunc -> func()
func (s *SignalManager) AddWatcher(sigType os.Signal, handlerTag string, handler SignalHandlerFunc) {
	if handler == nil {
		return
	}
	sig, ok := s.sigs[sigType]
	if ok {
		//覆写
		sig.handlers[handlerTag] = s.SignalHandlerFuncWrapper(sig.sigCh, handler)
		return
	}
	ch := make(chan os.Signal, 3)
	signal.Notify(ch, sigType)
	ss := &Signal{
		handlers: make(map[string]SignalHandlerFunc),
		sigCh:    ch,
	}
	ss.handlers[handlerTag] = s.SignalHandlerFuncWrapper(ch, handler)

	s.sigs[sigType] = ss
}

func (s *SignalManager) DeleteWatcher(sigType os.Signal, handlerTag string) {
	sig, ok := s.sigs[sigType]
	if !ok {
		return
	}
	if len(sig.handlers)-1 == 0 {
		signal.Stop(sig.sigCh)
		close(sig.sigCh)
		delete(s.sigs, sigType)
	}
	if _, ok := sig.handlers[handlerTag]; !ok {
		return
	}
	delete(sig.handlers, handlerTag)
}
