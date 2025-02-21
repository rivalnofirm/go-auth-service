package mail

import (
	"bytes"
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
	SendMailUpdatePassword(userId int64) error
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

	name := users.FirstName
	if users.LastName != "" {
		name = users.FirstName + " " + users.LastName
	}

	dataEmail := map[string]interface{}{
		"name":                name,
		"ip_address":          ipAddress,
		"user_agent":          userAgent,
		"login_time":          time.Now().Format("02 Jan 2006 15:04:05"),
		"reset_password_link": os.Getenv("URL_RESET_PASSWORD"),
	}

	file := os.Getenv("PATH_EMAIL_TEMPLATE") + "login.html"

	tmpl, err := template.ParseFiles(file)
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

func (uc *MailUseCase) SendMailUpdatePassword(userId int64) error {
	users, err := uc.RepoUser.GetUserDetailById(userId)
	if err != nil {
		return err
	}

	name := users.FirstName
	if users.LastName != "" {
		name = users.FirstName + " " + users.LastName
	}

	fileBody := os.Getenv("PATH_EMAIL_TEMPLATE") + "update-password.html"

	tmpl, err := template.ParseFiles(fileBody)
	if err != nil {
		return err
	}

	dataEmail := map[string]interface{}{
		"name": name,
	}

	var buffer bytes.Buffer

	if err = tmpl.Execute(&buffer, dataEmail); err != nil {
		return err
	}

	emailBody := buffer.String()

	err = helper.SendMail(users.Email, "Your Password Has Been Successfully Updated", emailBody)
	if err != nil {
		return err
	}

	return nil
}
