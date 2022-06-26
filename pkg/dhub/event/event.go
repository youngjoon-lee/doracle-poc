package event

import (
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

type HandlerFn func(ctypes.ResultEvent) error

type Event interface {
	Name() string
	Query() string
	Handler(ctypes.ResultEvent) error
}
