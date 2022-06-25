package dhub

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/youngjoon-lee/doracle-poc/pkg/dhub/event"
)

type Subscriber struct {
	client *rpchttp.HTTP
}

func StartSubscriber(rpcAddr string) (*Subscriber, error) {
	client, err := rpchttp.New(rpcAddr, "/websocket")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %v/websocket: %w", rpcAddr, err)
	}

	if err := client.Start(); err != nil {
		return nil, fmt.Errorf("failed to start subscriber: %w", err)
	}

	s := &Subscriber{
		client: client,
	}

	if err := s.subscribeAll(); err != nil {
		s.Stop()
		return nil, fmt.Errorf("failed to register all handlers: %w", err)
	}

	return s, nil
}

func (s *Subscriber) Stop() {
	log.Info("stopping subscriber...")
	s.client.Stop()
}

func (s *Subscriber) subscribeAll() error {
	for _, e := range event.Events() {
		if err := s.subscribe(e.Name(), e.Query(), e.Handler); err != nil {
			return fmt.Errorf("failed to subsribe: %v", e.Name())
		}
	}
	return nil
}

func (s *Subscriber) subscribe(subscriber, query string, handler func(ctypes.ResultEvent) error) error {
	resEventCh, err := s.client.Subscribe(context.Background(), subscriber, query)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	go func() {
		for resEvent := range resEventCh {
			log.Debugf("event detected: %v", resEvent)

			if err := handler(resEvent); err != nil {
				log.Errorf("failed to handle event: %v", err)
			}
		}
	}()

	log.Infof("subscription registered: %v / %v", subscriber, query)
	return nil
}
