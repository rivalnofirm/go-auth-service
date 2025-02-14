package user

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"regexp"
)

type UpdateUserProfileReqInterface interface {
	Validate() error
}

type UpdateUserProfileReq struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	BirthDate string `json:"birth_date"`
	Gender    string `json:"gender"`
}

func (dto *UpdateUserProfileReq) Validate() error {
	if err := validation.ValidateStruct(
		dto,
		validation.Field(
			&dto.FirstName,
			validation.Match(regexp.MustCompile(`^[a-zA-Z]+$`)).Error("must contain only alphabetic characters"),
		),
		validation.Field(
			&dto.LastName,
			validation.Match(regexp.MustCompile(`^[a-zA-Z]+$`)).Error("must contain only alphabetic characters"),
		), validation.Field(
			&dto.BirthDate,
			validation.Match(regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])$`)).Error("birth_date must be in 'YYYY-MM-DD' format"),
		),
		validation.Field(
			&dto.Gender,
			validation.In("male", "female").Error("gender must be either 'male' or 'female'"),
		),
	); err != nil {
		return err
	}
	return nil
}
