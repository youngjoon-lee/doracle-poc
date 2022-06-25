package event

import (
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

type Event interface {
	Name() string
	Query() string
	Handler(ctypes.ResultEvent) error
}

func Events() []Event {
	return []Event{
		joinEvent{},
	}
}
