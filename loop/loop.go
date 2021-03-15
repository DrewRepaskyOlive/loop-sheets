package loop

import (
	"context"

	ldk "github.com/open-olive/loop-development-kit/ldk/go"
)

const (
	loopName = "loop-sheets"
)

type Loop struct {
	ctx      context.Context
	cancel   context.CancelFunc
	logger   *ldk.Logger
	sidekick ldk.Sidekick
}

func Serve() error {
	log := ldk.NewLogger(loopName)
	loop, err := NewLoop(log)
	if err != nil {
		return err
	}
	ldk.ServeLoopPlugin(log, loop)
	return nil
}

func NewLoop(logger *ldk.Logger) (*Loop, error) {
	logger.Trace("NewLoop called: " + loopName)
	return &Loop{
		logger: logger,
	}, nil
}

func (l *Loop) LoopStart(sidekick ldk.Sidekick) error {
	l.logger.Trace("starting " + loopName)
	l.ctx, l.cancel = context.WithCancel(context.Background())
	l.sidekick = sidekick
	return nil
}

func (l *Loop) LoopStop() error {
	l.logger.Trace("stopping " + loopName)
	l.cancel()
	return nil
}
