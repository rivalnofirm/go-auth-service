package mail

import (
	"bytes"
	"context"
	"fmt"
	"go-auth-service/src/infra/constants/common"
	"html/template"
	"os"
	"time"

	"go-auth-service/src/infra/helper"
	repoHitory "go-auth-service/src/infra/persistence/postgres/history"
	repoUser "go-auth-service/src/infra/persistence/postgres/user"
	redis "go-auth-service/src/infra/persistence/redis/service"
)

type MailUCInterface interface {
	SendMailLogin(userId int64, ipAddress, userAgent string) error
	SendMailRegister(userId int64) error
}

type MailUseCase struct {
	Redis       redis.ServRedisInterface
	RepoUser    repoUser.UserRepository
	RepoHistory repoHitory.HistoryRepository
}

func NewMailUseCase(redisService redis.ServRedisInterface, repoUser repoUser.UserRepository, repoHistory repoHitory.HistoryRepository) *MailUseCase {
	return &MailUseCase{
		Redis:       redisService,
		RepoUser:    repoUser,
		RepoHistory: repoHistory,
	}
}

func (uc *MailUseCase) SendMailLogin(userId int64, ipAddress, userAgent string) error {
	users, err := uc.RepoUser.GetUserDetailById(userId)
	if err != nil {
		return err
	}

	history, err := uc.RepoHistory.GetByUserId(userId)
	if err != nil {
		return err
	}

	name := users.FirstName
	if users.LastName != "" {
		name = users.FirstName + " " + users.LastName
	}

	var fileBody string

	dataEmail := map[string]interface{}{
		"name":       name,
		"ip_address": ipAddress,
		"user_agent": userAgent,
		"login_time": time.Now().Format("02 Jan 2006 15:04:05"),
	}

	// Jika history kosong, gunakan fileBody default
	if len(history) == 0 {
		fileBody = os.Getenv("PATH_EMAIL_TEMPLATE") + "login.html"
	} else {
		found := false
		for _, h := range history {
			if h.UserAgent.Valid && h.UserAgent.String == userAgent {
				found = true
				break
			}
		}

		if !found {
			fileBody = os.Getenv("PATH_EMAIL_TEMPLATE") + "login-warning.html"

			emailEncrypt, _ := helper.Encrypt(users.Email)
			urlRevokeToke := os.Getenv("URL_API") + "/api/user/revoke-token/" + emailEncrypt
			dataEmail["url_revoke_token"] = urlRevokeToke

			revokeToken := fmt.Sprintf("%s:%s", common.RevokeTokenKey, users.Email)
			_ = uc.Redis.SetData(context.Background(), revokeToken, emailEncrypt, common.RevokeTokenExp)
		}
	}

	tmpl, err := template.ParseFiles(fileBody)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer

	if err = tmpl.Execute(&buffer, dataEmail); err != nil {
		return err
	}

	emailBody := buffer.String()

	err = helper.SendMail(users.Email, "New Login Detected on Your Account", emailBody)
	if err != nil {
		return err
	}

	_ = uc.RepoHistory.Create(userId, ipAddress, userAgent)

	return nil
}

func (uc *MailUseCase) SendMailRegister(userId int64) error {
	users, err := uc.RepoUser.GetUserDetailById(userId)
	if err != nil {
		return err
	}

	name := users.FirstName
	if users.LastName != "" {
		name = users.FirstName + " " + users.LastName
	}

	fileBody := os.Getenv("PATH_EMAIL_TEMPLATE") + "register.html"

	tmpl, err := template.ParseFiles(fileBody)
	if err != nil {
		return err
	}

	parsedTime, err := time.Parse("02-01-2006 15:04:05", users.CreatedAt)
	if err != nil {
		return err
	}
	formattedDate := parsedTime.Format("02 Jan 2006 15:04:05")

	dataEmail := map[string]interface{}{
		"name":              name,
		"email":             users.Email,
		"registration_date": formattedDate,
	}

	var buffer bytes.Buffer

	if err = tmpl.Execute(&buffer, dataEmail); err != nil {
		return err
	}

	emailBody := buffer.String()

	err = helper.SendMail(users.Email, "Your Account Has Been Successfully Created", emailBody)
	if err != nil {
		return err
	}

	return nil
}
