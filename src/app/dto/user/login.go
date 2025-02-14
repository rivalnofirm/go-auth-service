package user

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"regexp"
)

type LoginReqInterface interface {
	Validate() error
}

type LoginResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResp struct {
	AccessToken string `json:"access_token"`
}

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func (dto *LoginReq) Validate() error {
	return validation.ValidateStruct(
		dto,
		validation.Field(&dto.Email, validation.Required, validation.Match(emailRegex).Error("invalid email format")),
		validation.Field(&dto.Password, validation.Required),
	)
}
