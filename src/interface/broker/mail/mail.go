package mail

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/nats-io/nats.go"

	dtoNats "go-auth-service/src/app/dto/broker"
	uCMail "go-auth-service/src/app/usecases/mail"
	natsBroker "go-auth-service/src/infra/broker/nats"
	"go-auth-service/src/infra/constants/common"
)

type AuthInterface interface {
	Init()
}

type AuthImpl struct {
	Nats        *natsBroker.Nats
	UseCaseMail uCMail.MailUCInterface
}

func NewAuthWorker(nats *natsBroker.Nats, useCaseMail uCMail.MailUCInterface) AuthInterface {
	workerImpl := &AuthImpl{
		Nats:        nats,
		UseCaseMail: useCaseMail,
	}

	if nats.Status {
		workerImpl.Init()
	}

	return workerImpl
}

func (p *AuthImpl) Init() {
	var sendMailConcurrencyEnv = os.Getenv("SEND_MAIL_CONCURRENCY")
	sendMailConcurrency, _ := strconv.Atoi(sendMailConcurrencyEnv)

	for i := 0; i < sendMailConcurrency; i++ {
		go authWorker(i, p)
	}

}

func authWorker(concurrency int, w *AuthImpl) {
	subject := common.NatsAuthSubject
	queue := common.NatsAuthQueue

	_, err := w.Nats.Conn.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		dataConsume := dtoNats.AuthBrokerDto{}
		err := json.Unmarshal(msg.Data, &dataConsume)
		if err != nil {
			log.Println("[ERROR] unmarshal json err:", err)
			return
		}

		if dataConsume.Event == common.EventLogin {
			err = w.UseCaseMail.SendMailLogin(dataConsume.UserId, dataConsume.IpAddress, dataConsume.Device)
			if err != nil {
				log.Println("[ERROR] send mail err:", err)
			}
		} else if dataConsume.Event == common.EventRegister {
			err = w.UseCaseMail.SendMailRegister(dataConsume.UserId)
			if err != nil {
				log.Println("[ERROR] send mail err:", err)
			}
		}
	})

	if err != nil {
		log.Println("[ERROR] send mail err:", err)
	}

	log.Printf("Listening on [%s] at worker number [%d]", subject, concurrency)
}
