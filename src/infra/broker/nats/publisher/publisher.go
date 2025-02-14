package publisher

import (
	"log"

	"go-auth-service/src/infra/broker/nats"
)

type PublisherInterface interface {
	Nats(data []byte, subject string) error
}

type PublisherImpl struct {
	nats *nats.Nats
}

func NewPublisher(Nats *nats.Nats) PublisherInterface {
	natsPublisherImpl := &PublisherImpl{
		nats: Nats,
	}

	return natsPublisherImpl
}

func (p *PublisherImpl) Nats(data []byte, subject string) error {
	err := p.nats.Conn.Publish(subject, data)
	if err != nil {
		return err
	}

	err = p.nats.Conn.Flush()
	if err != nil {
		return err
	}

	if err := p.nats.Conn.LastError(); err != nil {
		return err
	}

	log.Printf("Published to [%s]\n", subject)

	return nil
}
