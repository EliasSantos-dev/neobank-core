package events

import (
	"errors"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

var errNotReady = errors.New("events: servidor NATS embutido não ficou pronto")

type Bus struct {
	srv *server.Server
	nc  *nats.Conn
}

// NewBus sobe um servidor NATS embutido (sem porta) e conecta in-process.
func NewBus() (*Bus, error) {
	srv, err := server.NewServer(&server.Options{DontListen: true})
	if err != nil {
		return nil, err
	}
	go srv.Start()
	if !srv.ReadyForConnections(5 * time.Second) {
		srv.Shutdown()
		return nil, errNotReady
	}
	nc, err := nats.Connect("", nats.InProcessServer(srv))
	if err != nil {
		srv.Shutdown()
		return nil, err
	}
	return &Bus{srv: srv, nc: nc}, nil
}

func (b *Bus) Publish(topic string, payload []byte) error {
	return b.nc.Publish(topic, payload)
}

func (b *Bus) Subscribe(topic string, handler func(topic string, payload []byte)) error {
	if _, err := b.nc.Subscribe(topic, func(m *nats.Msg) {
		handler(m.Subject, m.Data)
	}); err != nil {
		return err
	}
	return b.nc.Flush()
}

func (b *Bus) Close() {
	if b.nc != nil {
		_ = b.nc.Drain()
	}
	if b.srv != nil {
		b.srv.Shutdown()
	}
}
