package event

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
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

func (s *Subscriber) Subscribe(ev Event) error {
	resEventCh, err := s.client.Subscribe(context.Background(), ev.Name(), ev.Query())
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	go func() {
		for resEvent := range resEventCh {
			log.Debugf("event detected: %v", resEvent)

			if err := ev.Handler(resEvent); err != nil {
				log.Errorf("failed to handle event: %v", err)
			}
		}
	}()

	log.Infof("subscription registered: %v / %v", ev.Name(), ev.Query())
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
