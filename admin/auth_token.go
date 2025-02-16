package admin

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/authtoken"
)

// AuthToken is the interface package admin uses to provide a consolidated view of a token string and its DB model.
type AuthToken interface {
	Token() *authtoken.Token
	TokenModel() any
	OwnerID() string
}

// userAuthToken implements AuthToken for tokens belonging to a user.
type userAuthToken struct {
	model *database.UserAuthToken
	token *authtoken.Token
}

func (t *userAuthToken) Token() *authtoken.Token {
	return t.token
}

func (t *userAuthToken) TokenModel() any {
	return t.model
}

func (t *userAuthToken) OwnerID() string {
	if t.model.RepresentingUserID != nil {
		return *t.model.RepresentingUserID
	}

	return t.model.UserID
}

// IssueUserAuthToken generates and persists a new auth token for a user.
func (s *Service) IssueUserAuthToken(ctx context.Context, userID, clientID, displayName string, representingUserID *string, ttl *time.Duration) (AuthToken, error) {
	tkn := authtoken.NewRandom(authtoken.TypeUser)

	var expiresOn *time.Time
	if ttl != nil {
		t := time.Now().Add(*ttl)
		expiresOn = &t
	}

	uat, err := s.DB.InsertUserAuthToken(ctx, &database.InsertUserAuthTokenOptions{
		ID:                 tkn.ID.String(),
		SecretHash:         tkn.SecretHash(),
		UserID:             userID,
		AuthClientID:       &clientID,
		DisplayName:        displayName,
		RepresentingUserID: representingUserID,
		ExpiresOn:          expiresOn,
	})
	if err != nil {
		return nil, err
	}

	return &userAuthToken{model: uat, token: tkn}, nil
}

// serviceAuthToken implements AuthToken for tokens belonging to a service.
type serviceAuthToken struct {
	model *database.ServiceAuthToken
	token *authtoken.Token
}

func (t *serviceAuthToken) Token() *authtoken.Token {
	return t.token
}

func (t *serviceAuthToken) TokenModel() any {
	return t.model
}

func (t *serviceAuthToken) OwnerID() string {
	return t.model.ServiceID
}

// IssueServiceAuthToken generates and persists a new auth token for a service.
func (s *Service) IssueServiceAuthToken(ctx context.Context, serviceID string, ttl *time.Duration) (AuthToken, error) {
	tkn := authtoken.NewRandom(authtoken.TypeService)

	var expiresOn *time.Time
	if ttl != nil {
		t := time.Now().Add(*ttl)
		expiresOn = &t
	}

	sat, err := s.DB.InsertServiceAuthToken(ctx, &database.InsertServiceAuthTokenOptions{
		ID:         tkn.ID.String(),
		SecretHash: tkn.SecretHash(),
		ServiceID:  serviceID,
		ExpiresOn:  expiresOn,
	})
	if err != nil {
		return nil, err
	}

	return &serviceAuthToken{model: sat, token: tkn}, nil
}

// deploymentAuthToken implements AuthToken for tokens belonging to a deployment.
type deploymentAuthToken struct {
	model *database.DeploymentAuthToken
	token *authtoken.Token
}

func (t *deploymentAuthToken) Token() *authtoken.Token {
	return t.token
}

func (t *deploymentAuthToken) TokenModel() any {
	return t.model
}

func (t *deploymentAuthToken) OwnerID() string {
	return t.model.DeploymentID
}

// IssueDeploymentAuthToken generates and persists a new auth token for a deployment.
func (s *Service) IssueDeploymentAuthToken(ctx context.Context, deploymentID string, ttl *time.Duration) (AuthToken, error) {
	tkn := authtoken.NewRandom(authtoken.TypeDeployment)

	var expiresOn *time.Time
	if ttl != nil {
		t := time.Now().Add(*ttl)
		expiresOn = &t
	}

	dat, err := s.DB.InsertDeploymentAuthToken(ctx, &database.InsertDeploymentAuthTokenOptions{
		ID:           tkn.ID.String(),
		SecretHash:   tkn.SecretHash(),
		DeploymentID: deploymentID,
		ExpiresOn:    expiresOn,
	})
	if err != nil {
		return nil, err
	}

	return &deploymentAuthToken{model: dat, token: tkn}, nil
}

// magicAuthToken implements AuthToken for magic tokens belonging to a project.
type magicAuthToken struct {
	model *database.MagicAuthToken
	token *authtoken.Token
}

func (t *magicAuthToken) Token() *authtoken.Token {
	return t.token
}

func (t *magicAuthToken) TokenModel() any {
	return t.model
}

func (t *magicAuthToken) OwnerID() string {
	return t.model.ID
}

// IssueMagicAuthTokenOptions provides options for IssueMagicAuthToken.
type IssueMagicAuthTokenOptions struct {
	ProjectID       string
	TTL             *time.Duration
	CreatedByUserID *string
	Attributes      map[string]any
	FilterJSON      string
	Fields          []string
	State           string
	DisplayName     string
	Internal        bool
	Resources       []database.ResourceName
}

// IssueMagicAuthToken generates and persists a new magic auth token for a project.
func (s *Service) IssueMagicAuthToken(ctx context.Context, opts *IssueMagicAuthTokenOptions) (AuthToken, error) {
	tkn := authtoken.NewRandom(authtoken.TypeMagic)

	var expiresOn *time.Time
	if opts.TTL != nil {
		t := time.Now().Add(*opts.TTL)
		expiresOn = &t
	}

	dat, err := s.DB.InsertMagicAuthToken(ctx, &database.InsertMagicAuthTokenOptions{
		ID:              tkn.ID.String(),
		SecretHash:      tkn.SecretHash(),
		Secret:          tkn.Secret[:],
		ProjectID:       opts.ProjectID,
		ExpiresOn:       expiresOn,
		CreatedByUserID: opts.CreatedByUserID,
		Attributes:      opts.Attributes,
		FilterJSON:      opts.FilterJSON,
		Fields:          opts.Fields,
		State:           opts.State,
		DisplayName:     opts.DisplayName,
		Internal:        opts.Internal,
		Resources:       opts.Resources,
	})
	if err != nil {
		return nil, err
	}

	return &magicAuthToken{model: dat, token: tkn}, nil
}

// ValidateAuthToken validates an auth token against persistent storage.
func (s *Service) ValidateAuthToken(ctx context.Context, token string) (AuthToken, error) {
	parsed, err := authtoken.FromString(token)
	if err != nil {
		return nil, err
	}

	switch parsed.Type {
	case authtoken.TypeUser:
		uat, err := s.DB.FindUserAuthToken(ctx, parsed.ID.String())
		if err != nil {
			return nil, err
		}

		if uat.ExpiresOn != nil && uat.ExpiresOn.Before(time.Now()) {
			return nil, fmt.Errorf("auth token is expired")
		}

		if !bytes.Equal(uat.SecretHash, parsed.SecretHash()) {
			return nil, fmt.Errorf("invalid auth token")
		}

		s.Used.UserToken(uat.ID)
		s.Used.User(uat.UserID)

		return &userAuthToken{model: uat, token: parsed}, nil
	case authtoken.TypeService:
		sat, err := s.DB.FindServiceAuthToken(ctx, parsed.ID.String())
		if err != nil {
			return nil, err
		}

		if sat.ExpiresOn != nil && sat.ExpiresOn.Before(time.Now()) {
			return nil, fmt.Errorf("auth token is expired")
		}

		if !bytes.Equal(sat.SecretHash, parsed.SecretHash()) {
			return nil, fmt.Errorf("invalid auth token")
		}

		s.Used.ServiceToken(sat.ID)
		s.Used.Service(sat.ServiceID)

		return &serviceAuthToken{model: sat, token: parsed}, nil
	case authtoken.TypeDeployment:
		dat, err := s.DB.FindDeploymentAuthToken(ctx, parsed.ID.String())
		if err != nil {
			return nil, err
		}

		if dat.ExpiresOn != nil && dat.ExpiresOn.Before(time.Now()) {
			return nil, fmt.Errorf("auth token is expired")
		}

		if !bytes.Equal(dat.SecretHash, parsed.SecretHash()) {
			return nil, fmt.Errorf("invalid auth token")
		}

		s.Used.DeploymentToken(dat.ID)

		return &deploymentAuthToken{model: dat, token: parsed}, nil
	case authtoken.TypeMagic:
		mat, err := s.DB.FindMagicAuthToken(ctx, parsed.ID.String(), false)
		if err != nil {
			return nil, err
		}

		if mat.ExpiresOn != nil && mat.ExpiresOn.Before(time.Now()) {
			return nil, fmt.Errorf("auth token is expired")
		}

		if !bytes.Equal(mat.SecretHash, parsed.SecretHash()) {
			return nil, fmt.Errorf("invalid auth token")
		}

		s.Used.MagicAuthToken(mat.ID)

		return &magicAuthToken{model: mat, token: parsed}, nil
	default:
		return nil, fmt.Errorf("unknown auth token type %q", parsed.Type)
	}
}

// RevokeAuthToken removes an auth token from persistent storage.
func (s *Service) RevokeAuthToken(ctx context.Context, token string) error {
	parsed, err := authtoken.FromString(token)
	if err != nil {
		return err
	}
	switch parsed.Type {
	case authtoken.TypeUser:
		return s.DB.DeleteUserAuthToken(ctx, parsed.ID.String())
	case authtoken.TypeService:
		return s.DB.DeleteServiceAuthToken(ctx, parsed.ID.String())
	case authtoken.TypeDeployment:
		return fmt.Errorf("deployment auth tokens cannot be revoked")
	case authtoken.TypeMagic:
		return s.DB.DeleteMagicAuthToken(ctx, parsed.ID.String())
	default:
		return fmt.Errorf("unknown auth token type %q", parsed.Type)
	}
}
