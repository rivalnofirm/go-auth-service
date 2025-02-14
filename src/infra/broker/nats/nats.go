package nats

import (
	"log"
	"os"

	"github.com/nats-io/nats.go"
)

type Nats struct {
	Status bool
	Conn   *nats.Conn
}

func NewNats() *Nats {
	var Nats = new(Nats)
	statusEnv, ok := os.LookupEnv("NATS_STATUS")
	if ok {
		if statusEnv == "1" {
			Nats.Status = true
		}
	}

	if Nats.Status {
		NATSHost := nats.DefaultURL
		NATSHostEnv, ok := os.LookupEnv("NATS_HOST")
		if ok {
			NATSHost = NATSHostEnv
		}
		var err error
		Nats.Conn, err = nats.Connect(NATSHost)

		if err != nil {
			log.Panicf("Terjadi masalah koneksi pada NATS. %s\n", err.Error())
		}

		log.Println("connected to:", NATSHost)
	}

	return Nats
}
