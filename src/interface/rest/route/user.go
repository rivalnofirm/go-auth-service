package route

import (
	"github.com/go-chi/chi/v5"
	"net/http"

	handlersUser "go-auth-service/src/interface/rest/handlers/user"
)

func UserRouter(h handlersUser.UserHandlerInterface) http.Handler {
	r := chi.NewRouter()

	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Get("/me", h.Me)
	r.Get("/refresh-token", h.RefreshToken)
	r.Get("/logout", h.Logout)
	r.Get("/revoke-token/{email-encrypt}", h.RevokeToken)
	r.Put("/update-profile", h.UpdateProfile)
	r.Put("/update-profile-picture", h.UpdateProfilePicture)

	return r
}
