package oauth

import (
	"context"

	"github.com/garaekz/gonvelope/internal/config"
	"github.com/garaekz/gonvelope/internal/entity"
	"github.com/garaekz/gonvelope/pkg/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
)

// Service encapsulates usecase logic for oauth.
type Service interface {
	// GetAuthURL Generates the URL for the OAuth provider's consent page.
	GetAuthURL(provider string, state string) string
	// HandleCallback Exchanges the OAuth code for a token.
	HandleCallback(provider string, code string) (*oauth2.Token, error)
	// StoreAccount stores the user provider account
	StoreAccount(ctx context.Context, account entity.UserProviderAccount, provider string) error
}

type service struct {
	repo       Repository
	logger     log.Logger
	configs    *ProviderConfigs
	signingKey string
}

// NewService creates a new oauth service.
func NewService(repo Repository, logger log.Logger, configs *ProviderConfigs, signingKey string) Service {
	return service{repo, logger, configs, signingKey}
}

// ProviderConfigs represents the OAuth configurations for different providers.
type ProviderConfigs struct {
	Google  *config.OAuthConfig
	Outlook *config.OAuthConfig
}

// newOAuthConfig returns a new OAuth configuration for the given provider.
func (s service) newOAuthConfig(provider string) *oauth2.Config {
	var oauthConfig *oauth2.Config

	switch provider {
	case "google":
		oauthConfig = &oauth2.Config{
			ClientID:     s.configs.Google.ClientID,
			ClientSecret: s.configs.Google.ClientSecret,
			RedirectURL:  s.configs.Google.RedirectURL,
			Scopes:       s.configs.Google.Scopes,
			Endpoint:     google.Endpoint,
		}
	case "outlook":
		oauthConfig = &oauth2.Config{
			ClientID:     s.configs.Outlook.ClientID,
			ClientSecret: s.configs.Outlook.ClientSecret,
			RedirectURL:  s.configs.Outlook.RedirectURL,
			Scopes:       s.configs.Outlook.Scopes,
			Endpoint:     microsoft.AzureADEndpoint("common"),
		}
	}

	return oauthConfig
}

// GetAuthURL Generates the URL for the OAuth provider's consent page.
func (s service) GetAuthURL(provider string, state string) string {
	config := s.newOAuthConfig(provider)
	return config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// HandleCallback Exchanges the OAuth code for a token.
func (s service) HandleCallback(provider string, code string) (*oauth2.Token, error) {
	config := s.newOAuthConfig(provider)
	return config.Exchange(context.Background(), code)
}

// StoreAccount stores the user provider account
func (s service) StoreAccount(ctx context.Context, account entity.UserProviderAccount, name string) error {
	provider, err := s.repo.GetProviderByName(ctx, name)
	if err != nil {
		return err
	}
	account.ProviderID = provider.GetID()

	return s.repo.StoreUserProviderAccount(ctx, account)
}
