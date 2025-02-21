package common

import "time"

const (
	IsMasterDb    = true
	NotIsMasterDb = false

	SuperAdmin = 1
	Admin      = 2
	User       = 3

	AttachmentSizeLimit int64 = 10 * 1024 * 1024 // 10 MB

	EncryptKey = "a9B2cD3eF4gH5iJ6kL7mN8oP9qR0sT13"

	JwtKey        = "secret_key"
	JwtRefreshKey = "refresh_secret_key"

	AccessTokenExp  = 120 * time.Minute
	RefreshTokenExp = 7 * 24 * time.Hour
	UserDetailExp   = 24 * time.Hour
	RateLimit       = 5 * time.Minute
	RevokeTokenExp  = 30 * time.Minute

	NatsAuthSubject = `AuthSubject`
	NatsAuthQueue   = `AuthQueue`

	EventLogin          = "Login"
	EventRegister       = "Register"
	EventUpdatePassword = "UpdatePassword"

	// Redis Key
	LoginKey        = "login_attempt"
	AccessTokenKey  = "access_token"
	RefreshTokenKey = "refresh_token"
	UserIdKey       = "user_id"
	RevokeTokenKey  = "revoke_token"

	// Logout Reason
	Token_Revoked = "Token Revoked"
	User_Logout   = "User Logout"
)
