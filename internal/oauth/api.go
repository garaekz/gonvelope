package oauth

import (
	"github.com/garaekz/gonvelope/internal/auth"
	"github.com/garaekz/gonvelope/internal/entity"
	"github.com/garaekz/gonvelope/internal/errors"
	"github.com/garaekz/gonvelope/pkg/log"
	routing "github.com/garaekz/ozzo-routing"
	"github.com/gorilla/sessions"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *routing.RouteGroup, service Service, authHandler routing.Handler, logger log.Logger, store *sessions.CookieStore) {
	res := resource{service, logger, store}
	r.Get("google/callback", res.googleCallback)
	r.Use(authHandler)
	r.Get("google/login", res.googleLogin)
	r.Post("google/token", res.googleCallback)
	// r.Get("/outlook/login", res.outlookLogin)
	// r.Get("/outlook/callback", res.outlookCallback)
}

type resource struct {
	service Service
	logger  log.Logger
	store   *sessions.CookieStore
}

func (r resource) googleLogin(c *routing.Context) error {
	state := entity.GenerateID()

	session, err := r.store.Get(c.Request, "oauth")
	if err != nil {
		return errors.InternalServerError("Failed to get session")
	}
	session.Values["state"] = state
	if err := session.Save(c.Request, c.Response); err != nil {
		return errors.InternalServerError("Failed to save session")
	}

	authURL := r.service.GetAuthURL("google", state)
	// http.Redirect(c.Response, c.Request, authURL, http.StatusFound)
	return c.Write(struct {
		URL string `json:"auth_url"`
	}{authURL})
}

// TODO: This will be renamed and request will be have more fields
func (r resource) googleCallback(c *routing.Context) error {
	identity := auth.CurrentUser(c.Request.Context())
	userID := identity.GetID()

	var req struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}

	if err := c.Read(&req); err != nil {
		// logger.With(c.Request.Context()).Errorf("invalid request: %v", err)
		return errors.BadRequest("")
	}
	token, err := r.service.HandleCallback("google", req.Code)
	if err != nil {
		return errors.InternalServerError("Failed to get token with given code")
	}

	account := entity.UserProviderAccount{
		UserID:       userID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenExpiry:  token.Expiry,
	}

	err = r.service.StoreAccount(c.Request.Context(), account, "google")
	if err != nil {
		return errors.InternalServerError("Failed to store user provider account")
	}

	return c.Write(struct {
		Message string `json:"message"`
	}{
		Message: "Successfully linked Google account",
	})
}
