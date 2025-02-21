package user

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"regexp"
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
		validation.Field(
			&dto.Password,
			validation.Required,
			validation.Length(8, 20).Error("password must be between 8 and 20 characters"),
			validation.Match(regexp.MustCompile(`[A-Z]`)).Error("password must contain at least one uppercase letter"),
			validation.Match(regexp.MustCompile(`[a-z]`)).Error("password must contain at least one lowercase letter"),
			validation.Match(regexp.MustCompile(`[0-9]`)).Error("password must contain at least one number"),
			validation.Match(regexp.MustCompile(`[\W_]+`)).Error("password must contain at least one special character (e.g., @, #, $, %, etc.)"),
		),
	); err != nil {
		return err
	}
	return nil
}
