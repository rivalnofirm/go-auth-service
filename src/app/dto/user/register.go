package user

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation"
)

type RegisterReqInterface interface {
	Validate() error
}

type RegisterReq struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func (dto *RegisterReq) Validate() error {
	if err := validation.ValidateStruct(
		dto,
		validation.Field(
			&dto.FirstName,
			validation.Required,
			validation.Match(regexp.MustCompile(`^[a-zA-Z]+$`)).Error("must contain only alphabetic characters"),
		),
		validation.Field(
			&dto.LastName,
			validation.Required,
			validation.Match(regexp.MustCompile(`^[a-zA-Z]+$`)).Error("must contain only alphabetic characters"),
		),
		validation.Field(
			&dto.Email,
			validation.Required,
			validation.Match(regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)).Error("invalid email format"),
		),
		validation.Field(&dto.Password, validation.Required),
	); err != nil {
		return err
	}
	return nil
}
