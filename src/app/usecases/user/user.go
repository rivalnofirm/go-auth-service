package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"time"

	dtoNats "go-auth-service/src/app/dto/broker"
	"go-auth-service/src/app/dto/user"
	natsPublisher "go-auth-service/src/infra/broker/nats/publisher"
	"go-auth-service/src/infra/constants/common"
	errorMessage "go-auth-service/src/infra/constants/error_message"
	"go-auth-service/src/infra/helper"
	"go-auth-service/src/infra/models"
	repoHistory "go-auth-service/src/infra/persistence/postgres/history"
	repoUser "go-auth-service/src/infra/persistence/postgres/user"
	redis "go-auth-service/src/infra/persistence/redis/service"
)

type UserUCInterface interface {
	Register(data *user.RegisterReq) error
	Login(data *user.LoginReq, ipAddress, userAgent string) (*user.LoginResp, error)
	VerifyToken(userId int64) (*user.UserDetails, error)
	RefreshToken(refreshToken string) (*user.RefreshTokenResp, error)
	Logout(userId int64, email, userAgent string) error
	RevokeToken(emailEncrypt string) error
	UpdateUserProfile(userId int64, data *user.UpdateUserProfileReq) error
	UpdateProfilePicture(userId int64, fileHeader *multipart.FileHeader) error
}

type userUseCase struct {
	NatsPublisher natsPublisher.PublisherInterface
	Redis         redis.ServRedisInterface
	RepoUser      repoUser.UserRepository
	RepoHistory   repoHistory.HistoryRepository
}

func NewUserUseCase(
	natsPublisher natsPublisher.PublisherInterface,
	redisService redis.ServRedisInterface,
	repoUser repoUser.UserRepository,
	repoHistory repoHistory.HistoryRepository,
) UserUCInterface {
	return &userUseCase{
		NatsPublisher: natsPublisher,
		Redis:         redisService,
		RepoUser:      repoUser,
		RepoHistory:   repoHistory,
	}
}

func (uc *userUseCase) Register(data *user.RegisterReq) error {
	_, err := uc.RepoUser.GetByEmail(data.Email)
	if err == nil {
		return errors.New(errorMessage.EmailAlready)
	}

	userId, err := uc.RepoUser.Create(data)
	if err != nil {
		log.Println(err)
		return err
	}

	sendMailDto := dtoNats.AuthBrokerDto{
		UserId: userId,
		Event:  common.EventRegister,
	}

	dataPublishMarshal, _ := json.Marshal(sendMailDto)
	err = uc.NatsPublisher.Nats(dataPublishMarshal, common.NatsAuthQueue)
	if err != nil {
		return err
	}

	return nil
}

func (uc *userUseCase) Login(data *user.LoginReq, ipAddress, userAgent string) (*user.LoginResp, error) {
	var err error
	var resp user.LoginResp
	var users *models.User

	// Rate limiting
	loginKey := fmt.Sprintf("%s:%s", common.LoginKey, data.Email)
	allowed, _ := uc.Redis.IsAllowed(context.Background(), loginKey, 5, common.RateLimit)
	if !allowed {
		return nil, fmt.Errorf(errorMessage.ToManyRequest)
	}

	accessTokenKey := fmt.Sprintf("%s:%s", common.AccessTokenKey, data.Email)
	refreshTokenKey := fmt.Sprintf("%s:%s", common.RefreshTokenKey, data.Email)

	cachedAccessToken, err := uc.Redis.GetData(context.Background(), accessTokenKey)
	if err == nil && cachedAccessToken != "" {
		claims, err := helper.VerifyToken(cachedAccessToken)
		if err == nil {
			resp.AccessToken = cachedAccessToken
			resp.RefreshToken, _ = uc.Redis.GetData(context.Background(), refreshTokenKey)

			users = &models.User{
				Id:    claims.UserID,
				Email: claims.Email,
			}
		} else {
			_ = uc.Redis.DeleteData(context.Background(), accessTokenKey)
		}
	}

	if users == nil {
		cachedRefreshToken, err := uc.Redis.GetData(context.Background(), refreshTokenKey)
		if err == nil && cachedRefreshToken != "" {
			claims, err := helper.VerifyRefreshToken(cachedRefreshToken)
			if err == nil {
				users = &models.User{
					Id:    claims.UserID,
					Email: claims.Email,
				}

				resp.AccessToken, err = helper.GenerateToken(users)
				if err != nil {
					return nil, err
				}

				resp.RefreshToken = cachedRefreshToken

				_ = uc.Redis.SetData(context.Background(), accessTokenKey, resp.AccessToken, common.AccessTokenExp)
			} else {
				_ = uc.Redis.DeleteData(context.Background(), refreshTokenKey)
			}
		}
	}

	if users == nil {
		users, err = uc.RepoUser.GetByEmail(data.Email)
		if err != nil {
			return nil, err
		}

		if err = helper.VerifyPassword(users.Password, data.Password); err != nil {
			return nil, fmt.Errorf(errorMessage.InvalidPassword)
		}

		resp.AccessToken, err = helper.GenerateToken(users)
		if err != nil {
			return nil, err
		}

		resp.RefreshToken, err = helper.GenerateRefreshToken(users)
		if err != nil {
			return nil, err
		}

		_ = uc.Redis.SetData(context.Background(), accessTokenKey, resp.AccessToken, common.AccessTokenExp)
		_ = uc.Redis.SetData(context.Background(), refreshTokenKey, resp.RefreshToken, common.RefreshTokenExp)
	}

	sendMailDto := dtoNats.AuthBrokerDto{
		UserId:    users.Id,
		IpAddress: ipAddress,
		Device:    userAgent,
		Event:     common.EventLogin,
	}

	dataPublishMarshal, _ := json.Marshal(sendMailDto)
	err = uc.NatsPublisher.Nats(dataPublishMarshal, common.NatsAuthSubject)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (uc *userUseCase) VerifyToken(userId int64) (*user.UserDetails, error) {
	userKey := fmt.Sprintf("%s:%d", common.UserIdKey, userId)
	userData, err := uc.Redis.GetData(context.Background(), userKey)
	if err == nil && userData != "" {
		result := &user.UserDetails{}
		err = json.Unmarshal([]byte(userData), result)
		if err == nil {
			return result, nil
		}
	}

	result, err := uc.RepoUser.GetUserDetailById(userId)
	if err != nil {
		return nil, err
	}

	dataRedis, _ := json.Marshal(result)
	err = uc.Redis.SetData(context.Background(), userKey, dataRedis, common.UserDetailExp)
	if err != nil {
		log.Println("Failed to save user details to Redis", err)
	}

	return result, nil
}

func (uc *userUseCase) RefreshToken(refreshToken string) (*user.RefreshTokenResp, error) {
	var resp user.RefreshTokenResp
	var users *models.User

	claims, err := helper.VerifyRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	userKey := fmt.Sprintf("%s:%d", common.UserIdKey, claims.UserID)
	userData, err := uc.Redis.GetData(context.Background(), userKey)
	if err == nil && userData != "" {
		_ = json.Unmarshal([]byte(userData), &users)
	} else {
		usersFromDB, err := uc.RepoUser.GetById(claims.UserID)
		if err != nil {
			return nil, err
		}
		users = usersFromDB
	}

	accessToken, err := helper.GenerateToken(users)
	if err != nil {
		return nil, err
	}

	accessTokenKey := fmt.Sprintf("%s:%s", common.AccessTokenKey, claims.Email)
	_ = uc.Redis.DeleteData(context.Background(), accessToken)
	_ = uc.Redis.SetData(context.Background(), accessTokenKey, accessToken, common.AccessTokenExp)

	resp.AccessToken = accessToken

	return &resp, nil
}

func (uc *userUseCase) Logout(userId int64, email, userAgent string) error {
	err := uc.RepoHistory.UpdateLogoutByUserIdAndUserAgent(userId, common.User_Logout, userAgent)
	if err != nil {
		return err
	}

	accessTokenKey := fmt.Sprintf("%s:%s", common.AccessTokenKey, email)
	refreshTokenKey := fmt.Sprintf("%s:%s", common.RefreshTokenKey, email)

	_ = uc.Redis.DeleteData(context.Background(), accessTokenKey)
	_ = uc.Redis.DeleteData(context.Background(), refreshTokenKey)

	return nil
}

func (uc *userUseCase) RevokeToken(emailEncrypt string) error {
	email, err := helper.Decrypt(emailEncrypt)
	if err != nil {
		return err
	}

	revokeToken := fmt.Sprintf("%s:%s", common.RevokeTokenKey, email)
	cache, _ := uc.Redis.GetData(context.Background(), revokeToken)
	if cache == "" {
		return errors.New(errorMessage.ExpiredToken)
	}

	users, err := uc.RepoUser.GetByEmail(email)
	if err != nil {
		return err
	}

	err = uc.RepoHistory.UpdateLogoutByUserId(users.Id, common.Token_Revoked)
	if err != nil {
		return err
	}

	accessTokenKey := fmt.Sprintf("%s:%s", common.AccessTokenKey, email)
	refreshTokenKey := fmt.Sprintf("%s:%s", common.RefreshTokenKey, email)

	_ = uc.Redis.DeleteData(context.Background(), accessTokenKey)
	_ = uc.Redis.DeleteData(context.Background(), refreshTokenKey)
	_ = uc.Redis.DeleteData(context.Background(), revokeToken)

	return nil
}

func (uc *userUseCase) UpdateUserProfile(userId int64, data *user.UpdateUserProfileReq) error {
	var users *user.UserDetails
	userKey := fmt.Sprintf("%s:%d", common.UserIdKey, userId)

	userData, err := uc.Redis.GetData(context.Background(), userKey)
	if err == nil && userData != "" {
		users = &user.UserDetails{}
		if err := json.Unmarshal([]byte(userData), users); err != nil {
			users = nil
		}
	}

	if users == nil {
		users, err = uc.RepoUser.GetUserDetailById(userId)
		if err != nil {
			return err
		}
	}

	if data.FirstName == "" {
		data.FirstName = users.FirstName
	}

	if data.LastName == "" {
		data.LastName = users.LastName
	}

	if data.BirthDate == "" {
		data.BirthDate = users.BirthDate
	} else {
		if _, err := time.Parse("2006-01-02", data.BirthDate); err != nil {
			return errors.New(errorMessage.RequestPayload)
		}
	}

	if data.Gender == "" {
		data.Gender = users.Gender
	} else {
		if data.Gender != "male" && data.Gender != "female" {
			return errors.New(errorMessage.RequestPayload)
		}
	}

	err = uc.RepoUser.UpdateProfileByUserId(userId, data.FirstName, data.LastName, data.BirthDate, data.Gender)
	if err != nil {
		return err
	}

	_ = uc.Redis.DeleteData(context.Background(), userKey)

	return nil
}

func (uc *userUseCase) UpdateProfilePicture(userId int64, fileHeader *multipart.FileHeader) error {
	var users *user.UserDetails
	userKey := fmt.Sprintf("%s:%d", common.UserIdKey, userId)

	userData, err := uc.Redis.GetData(context.Background(), userKey)
	if err == nil && userData != "" {
		users = &user.UserDetails{}
		if err := json.Unmarshal([]byte(userData), users); err != nil {
			users = nil
		}
	}

	if users == nil {
		users, err = uc.RepoUser.GetUserDetailById(userId)
		if err != nil {
			return err
		}
	}

	path, err := helper.UploadPicture(fileHeader, users.UserId)
	if err != nil {
		return err
	}

	err = uc.RepoUser.UpdateProfilePictureByUserId(users.UserId, path)
	if err != nil {
		return err
	}

	_ = uc.Redis.DeleteData(context.Background(), userKey)

	return nil
}
