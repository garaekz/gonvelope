package auth

import (
	"github.com/garaekz/gonvelope/internal/errors"
	"github.com/garaekz/gonvelope/pkg/log"
	routing "github.com/garaekz/ozzo-routing"
)

// RegisterHandlers registers handlers for different HTTP requests.
func RegisterHandlers(rg *routing.RouteGroup, service Service, logger log.Logger) {
	rg.Post("login", login(service, logger))
	rg.Post("register", register(service, logger))
}

// login returns a handler that handles user login request.
func login(service Service, logger log.Logger) routing.Handler {
	return func(c *routing.Context) error {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.Read(&req); err != nil {
			logger.With(c.Request.Context()).Errorf("invalid request: %v", err)
			return errors.BadRequest("")
		}

		token, err := service.Login(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			return err
		}
		return c.Write(struct {
			Token string `json:"token"`
		}{token})
	}
}

// register returns a handler that handles user registration request.
func register(service Service, logger log.Logger) routing.Handler {
	return func(c *routing.Context) error {
		var req struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.Read(&req); err != nil {
			logger.With(c.Request.Context()).Errorf("invalid request: %v", err)
			return errors.BadRequest("")
		}

		if err := service.Register(c.Request.Context(), req.Name, req.Email, req.Password); err != nil {
			return err
		}
		return c.WriteWithStatus(struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
		}{
			Status:  201,
			Message: "User was created successfully",
		}, 201)
	}
}
