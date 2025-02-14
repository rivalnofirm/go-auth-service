package user

import (
	"encoding/json"
	"go-auth-service/src/infra/helper"
	"log"
	"net/http"
	"strings"

	"github.com/lib/pq"
	"go-auth-service/src/app/dto/user"
	usecases "go-auth-service/src/app/usecases/user"
	errorMessage "go-auth-service/src/infra/constants/error_message"
	"go-auth-service/src/interface/rest/response"
)

type UserHandlerInterface interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	VerifyToken(w http.ResponseWriter, r *http.Request)
	RefreshToken(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	RevokeToken(w http.ResponseWriter, r *http.Request)
	UpdateProfile(w http.ResponseWriter, r *http.Request)
}

type userHandler struct {
	usecase usecases.UserUCInterface
}

func NewUserHandler(h usecases.UserUCInterface) UserHandlerInterface {
	return &userHandler{
		usecase: h,
	}
}

func (h *userHandler) RevokeToken(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")

	if len(pathParts) < 5 {
		response.JSON(w, http.StatusBadRequest, "error", errorMessage.BadRequest, nil)
		return
	}

	emailEncrypt := pathParts[4]

	err := h.usecase.RevokeToken(emailEncrypt)
	if err != nil {
		response.JSON(w, http.StatusUnauthorized, "error", errorMessage.ExpiredToken, nil)
	}

	http.Redirect(w, r, "https://www.kemhan.go.id/", http.StatusFound)

}

func (h *userHandler) Register(w http.ResponseWriter, r *http.Request) {
	postDTO := user.RegisterReq{}
	err := json.NewDecoder(r.Body).Decode(&postDTO)
	if err != nil {
		log.Println(err)
		response.JSON(w, http.StatusBadRequest, "error", errorMessage.RequestPayload, nil)
		return
	}

	err = postDTO.Validate()
	if err != nil {
		log.Println(err)
		response.JSON(w, http.StatusBadRequest, "error", err.Error(), nil)
		return
	}

	err = h.usecase.Register(&postDTO)
	if err != nil {
		log.Println(err)
		if err.Error() == errorMessage.EmailAlready {
			response.JSON(w, http.StatusConflict, "error", errorMessage.EmailAlready, nil)
			return
		}

		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			response.JSON(w, http.StatusConflict, "error", pqErr.Error(), nil)
			return
		}

		response.JSON(w, http.StatusConflict, "error", errorMessage.FailedCreateData, nil)
		return
	}

	response.JSON(w, http.StatusOK, "success", "successful register user", nil)
}

func (h *userHandler) Login(w http.ResponseWriter, r *http.Request) {
	userIp := helper.GetRealIP(r)
	userAgent := r.Header.Get("User-Agent")

	postDTO := user.LoginReq{}
	err := json.NewDecoder(r.Body).Decode(&postDTO)
	if err != nil {
		log.Println(err)
		response.JSON(w, http.StatusBadRequest, "error", errorMessage.RequestPayload, nil)
		return
	}

	err = postDTO.Validate()
	if err != nil {
		log.Println(err)
		response.JSON(w, http.StatusBadRequest, "error", err.Error(), nil)
		return
	}

	token, err := h.usecase.Login(&postDTO, userIp, userAgent)
	if err != nil {
		log.Println(err)
		response.JSON(w, http.StatusUnauthorized, "error", errorMessage.Unauthorized, nil)
		return
	}

	response.JSON(w, http.StatusOK, "success", "successful login", token)
}

func (h *userHandler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		response.JSON(w, http.StatusUnauthorized, "error", errorMessage.MissingToken, nil)
		return
	}

	claims, err := helper.VerifyToken(token)
	if err != nil {
		log.Println(err)
		response.JSON(w, http.StatusUnauthorized, "error", errorMessage.InvalidToken, nil)
		return
	}

	userDetail, err := h.usecase.VerifyToken(claims.UserID)
	if err != nil {
		log.Println(err)
		response.JSON(w, http.StatusUnauthorized, "error", err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, "success", "token is valid", userDetail)
}

func (h *userHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.Header.Get("Authorization")
	if refreshToken == "" {
		response.JSON(w, http.StatusUnauthorized, "error", errorMessage.MissingToken, nil)
		return
	}

	accessToken, err := h.usecase.RefreshToken(refreshToken)
	if err != nil {
		log.Println(err)
		response.JSON(w, http.StatusUnauthorized, "error", errorMessage.Unauthorized, nil)
		return
	}

	response.JSON(w, http.StatusOK, "success", "refresh token is valid", accessToken)
}

func (h *userHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userAgent := r.Header.Get("User-Agent")

	token := r.Header.Get("Authorization")
	if token == "" {
		response.JSON(w, http.StatusUnauthorized, "error", errorMessage.MissingToken, nil)
		return
	}

	claims, err := helper.VerifyToken(token)
	if err != nil {
		log.Println(err)
		response.JSON(w, http.StatusUnauthorized, "error", errorMessage.InvalidToken, nil)
		return
	}

	err = h.usecase.Logout(claims.UserID, claims.Email, userAgent)
	if err != nil {
		log.Println(err)
		response.JSON(w, http.StatusUnauthorized, "error", errorMessage.Unauthorized, nil)
		return
	}

	response.JSON(w, http.StatusOK, "success", "logout", nil)

}

func (h *userHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		response.JSON(w, http.StatusUnauthorized, "error", errorMessage.MissingToken, nil)
		return
	}

	claims, err := helper.VerifyToken(token)
	if err != nil {
		log.Println(err)
		response.JSON(w, http.StatusUnauthorized, "error", errorMessage.InvalidToken, nil)
		return
	}

	postDTO := user.UpdateUserProfileReq{}
	err = json.NewDecoder(r.Body).Decode(&postDTO)
	if err != nil {
		log.Println(err)
		response.JSON(w, http.StatusBadRequest, "error", errorMessage.RequestPayload, nil)
		return
	}

	err = postDTO.Validate()
	if err != nil {
		log.Println(err)
		response.JSON(w, http.StatusBadRequest, "error", err.Error(), nil)
		return
	}

	err = h.usecase.UpdateUserProfile(claims.UserID, &postDTO)
	if err != nil {
		log.Println(err)
		response.JSON(w, http.StatusUnauthorized, "error", err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, "success", "update profile", nil)
}
