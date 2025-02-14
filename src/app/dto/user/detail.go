package user

type UserDetails struct {
	UserId    int64  `json:"user_id"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	UserType  string `json:"user_type"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Picture   string `json:"picture"`
	BirthDate string `json:"birth_date"`
	Gender    string `json:"gender"`
	Verified  string `json:"verified"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	DeletedAt string `json:"deleted_at"`
}
