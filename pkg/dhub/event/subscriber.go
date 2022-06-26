package event

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	"github.com/youngjoon-lee/doracle-poc/pkg/app"
)

type Subscriber struct {
	client *rpchttp.HTTP
}

func NewSubscriber(rpcAddr string) (*Subscriber, error) {
	client, err := rpchttp.New(rpcAddr, "/websocket")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %v/websocket: %w", rpcAddr, err)
	}

	return &Subscriber{
		client: client,
	}, nil
}

func (s *Subscriber) Start() error {
	log.Info("starting subscriber...")
	return s.client.Start()
}

func (s *Subscriber) Stop() {
	log.Info("stopping subscriber...")
	s.client.Stop()
}

func (s *Subscriber) SubscribeAll(app *app.App) error {
	for _, e := range Events(app) {
		if err := s.subscribe(e.Name(), e.Query(), e.Handler); err != nil {
			return fmt.Errorf("failed to subsribe: %v", e.Name())
		}
	}
	return nil
}

func (s *Subscriber) subscribe(subscriber, query string, handler HandlerFn) error {
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

func (s *Subscriber) SubscribeOnce(ctx context.Context, ev Event) error {
	resEventCh, err := s.client.Subscribe(ctx, ev.Name(), ev.Query())
	if err != nil {
		return fmt.Errorf("failed to subscribe once: %w", err)
	}
	defer func() {
		if err := s.client.Unsubscribe(ctx, ev.Name(), ev.Query()); err != nil {
			log.Errorf("failed to unsubscribe: %v", err)
		}
	}()

	resEvent := <-resEventCh
	log.Debugf("event detected once: %v", resEvent)

	if err := ev.Handler(resEvent); err != nil {
		return fmt.Errorf("failed to handle event: %w", err)
	}

	return nil
}
