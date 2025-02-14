package usecases

import (
	mailUC "go-auth-service/src/app/usecases/mail"
	userUC "go-auth-service/src/app/usecases/user"
)

type AllUseCases struct {
	MailUC mailUC.MailUCInterface
	UserUC userUC.UserUCInterface
}
