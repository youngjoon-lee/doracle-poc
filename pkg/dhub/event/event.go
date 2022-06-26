package event

import (
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/youngjoon-lee/doracle-poc/pkg/app"
)

type HandlerFn func(ctypes.ResultEvent) error

type Event interface {
	Name() string
	Query() string
	Handler(ctypes.ResultEvent) error
}

func Events(app *app.App) []Event {
	return []Event{
		joinEvent{app: app},
	}
}
