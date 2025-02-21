package error_message

const (
	RequestPayload           = "invalid request payload"
	InvalidPassword          = "invalid password"
	InvalidToken             = "invalid token"
	ExpiredToken             = "token has expired"
	MissingToken             = "Token is missing or not found"
	UserNotFound             = "user not found"
	FailedUpdateData         = "failed to update data"
	EmailAlready             = "email already in use"
	PhoneAlready             = "phone already in use"
	FailedCreateData         = "failed to create data"
	Unauthorized             = "unauthorized, please check your account"
	ToManyRequest            = "too many requests"
	FailedDeleteData         = `failed to delete data`
	LoginHistoryNotFound     = "login history not found"
	BadRequest               = "bad request"
	UserRefreshTokenNotFound = "user refresh token not found"
	MissingUserAgent         = "missing user agent"
)
