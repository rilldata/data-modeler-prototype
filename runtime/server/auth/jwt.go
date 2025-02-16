package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/go-jose/go-jose/v3"
	"github.com/golang-jwt/jwt/v4"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

// Issuer creates JWTs with claims for an Audience.
// The Issuer is used by the admin server to create JWTs for the runtimes it manages based on a user's control-plane permissions.
type Issuer struct {
	issuerURL  string
	signingKey jose.JSONWebKey
	publicJWKS []byte
}

// Set issuerURL
func (i *Issuer) SetIssuerURL(issuerURL string) {
	i.issuerURL = issuerURL
}

// Get issuerURL
func (i *Issuer) GetIssuerURL() string {
	return i.issuerURL
}

// NewIssuer creates an issuer from a JWKS. The JWKS must contain private keys.
// The key identified by signingKeyID will be used to sign new JWTs.
func NewIssuer(issuerURL, signingKeyID string, jwksJSON []byte) (*Issuer, error) {
	// Parse the private JWKS
	var jwks jose.JSONWebKeySet
	err := json.Unmarshal(jwksJSON, &jwks)
	if err != nil {
		return nil, fmt.Errorf("invalid JWKS: %w", err)
	}

	// Extract signing key (must be a valid, private key)
	var signingKey jose.JSONWebKey
	for i := 0; i < len(jwks.Keys); i++ {
		if jwks.Keys[i].KeyID == signingKeyID {
			signingKey = jwks.Keys[i]
			break
		}
	}
	if !signingKey.Valid() || signingKey.IsPublic() {
		return nil, fmt.Errorf("invalid signing key %q", signingKeyID)
	}

	// Map JWKS to public keys and serialize to JSON
	var publicJWKS jose.JSONWebKeySet
	for i := 0; i < len(jwks.Keys); i++ {
		publicKey := jwks.Keys[i].Public()
		if !publicKey.Valid() {
			return nil, fmt.Errorf("invalid signing key in JWKS")
		}
		publicJWKS.Keys = append(publicJWKS.Keys, publicKey)
	}
	publicJSON, err := json.Marshal(publicJWKS)
	if err != nil {
		return nil, err
	}

	return &Issuer{
		issuerURL:  issuerURL,
		signingKey: signingKey,
		publicJWKS: publicJSON,
	}, nil
}

// NewEphemeralIssuer creates an Issuer using a generated JWKS.
// It is useful for development and testing, but should not be used in production.
func NewEphemeralIssuer(issuerURL string) (*Issuer, error) {
	// NOTE: JWKS generation based on: https://github.com/go-jose/go-jose/blob/v3/jose-util/generate.go

	// Generate RSA private key
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Create JWK
	jwk := jose.JSONWebKey{
		Key:       rsaKey,
		Algorithm: string(jose.RS256),
		Use:       "sig",
	}

	// Set key ID based on JWK thumbprint
	thumb, err := jwk.Thumbprint(crypto.SHA256)
	if err != nil {
		return nil, err
	}
	jwk.KeyID = base64.URLEncoding.EncodeToString(thumb)

	// Create JWKS JSON
	jwks := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}}
	jwksJSON, err := json.Marshal(jwks)
	if err != nil {
		return nil, err
	}

	return NewIssuer(issuerURL, jwk.KeyID, jwksJSON)
}

// TokenOptions provides options for Issuer.NewToken.
type TokenOptions struct {
	AudienceURL         string
	Subject             string
	TTL                 time.Duration
	SystemPermissions   []Permission
	InstancePermissions map[string][]Permission
	Attributes          map[string]any
	SecurityRules       []*runtimev1.SecurityRule
}

// NewToken issues a new JWT based on the provided options.
func (i *Issuer) NewToken(opts TokenOptions) (string, error) {
	// Since the security policy is a proto message, we need to serialize it using protojson instead of json.Marshal.
	var sec []json.RawMessage
	if len(opts.SecurityRules) > 0 {
		sec = make([]json.RawMessage, len(opts.SecurityRules))
		for i, rule := range opts.SecurityRules {
			data, err := protojson.Marshal(rule)
			if err != nil {
				return "", err
			}
			sec[i] = data
		}
	}

	// Create claims
	now := time.Now()
	claims := &jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(opts.TTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    i.issuerURL,
			Subject:   opts.Subject,
			Audience:  []string{opts.AudienceURL},
		},
		System:    opts.SystemPermissions,
		Instances: opts.InstancePermissions,
		Attrs:     opts.Attributes,
		Security:  sec,
	}

	// Create token
	token := jwt.NewWithClaims(jwt.GetSigningMethod(i.signingKey.Algorithm), claims)
	token.Header["kid"] = i.signingKey.KeyID
	res, err := token.SignedString(i.signingKey.Key)
	if err != nil {
		return "", err
	}

	return res, nil
}

// WellKnownHandler serves the public keys of the Issuer's JWKS.
// The Audience expects it to be mounted on {issuerURL}/.well-known/jwks.json.
func (i *Issuer) WellKnownHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(i.publicJWKS)
	})
}

// Audience represents a receiver of tokens from Issuer.
// The Audience is used by the runtime to parse claims from a JWT.
// It parses and validates tokens and resolves permissions.
// It refreshes its JWKS in the background from {issuerURL}/.well-known/jwks.json.
type Audience struct {
	issuerURL   string
	audienceURL string
	jwks        *keyfunc.JWKS
}

// OpenAudience creates an Audience. Remember to call Close() when done.
// The issuerURL should be the external URL of the issuing admin server.
// The issuerURL is expected to serve a JWKS on /.well-known/jwks.json.
// The audienceURL should be the external URL of the receiving runtime server.
func OpenAudience(ctx context.Context, logger *zap.Logger, issuerURL, audienceURL string) (*Audience, error) {
	// To be safe, require issuer and audience is provided
	if issuerURL == "" {
		return nil, fmt.Errorf("issuerURL is not set")
	}
	if audienceURL == "" {
		return nil, fmt.Errorf("audienceURL is not set")
	}

	// The JWKS is assumed to be served on issuer
	jwksURL, err := url.JoinPath(issuerURL, ".well-known/jwks.json")
	if err != nil {
		return nil, err
	}

	// Setup keyfunc that refreshes the JWKS in the background.
	// It returns an error if the initial fetch fails. So we wrap it with a retry in case the admin server is not ready.
	var jwks *keyfunc.JWKS
	for i := 0; i < 20; i++ {
		jwks, err = keyfunc.Get(jwksURL, keyfunc.Options{
			Ctx: ctx,
			RefreshErrorHandler: func(err error) {
				logger.Error("JWK refresh failed", zap.Error(err))
			},
			RefreshInterval:   time.Hour,
			RefreshRateLimit:  time.Minute * 5,
			RefreshTimeout:    time.Second * 10,
			RefreshUnknownKID: true,
		})
		if err != nil {
			logger.Info("JWKS fetch failed, retrying in 5s", zap.Error(err))
			select {
			case <-time.After(time.Second * 5):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}
	if err != nil {
		return nil, err
	}

	return &Audience{
		issuerURL:   issuerURL,
		audienceURL: audienceURL,
		jwks:        jwks,
	}, nil
}

// Close stops background refresh of the JWKS.
func (a *Audience) Close() {
	a.jwks.EndBackground()
}

// ParseAndValidate parses and validates a JWT and returns Claims if successful.
func (a *Audience) ParseAndValidate(tokenStr string) (Claims, error) {
	claims := &jwtClaims{}

	_, err := jwt.ParseWithClaims(tokenStr, claims, a.jwks.Keyfunc) // NOTE: also validates claims
	if err != nil {
		return nil, err
	}

	if !claims.VerifyIssuer(a.issuerURL, true) {
		return nil, fmt.Errorf("invalid token issuer %q (expected %q)", claims.Issuer, a.issuerURL)
	}

	if !claims.VerifyAudience(a.audienceURL, true) {
		return nil, fmt.Errorf("invalid token audience %q (expected %q)", claims.Audience, a.audienceURL)
	}

	return claims, nil
}

func NewDevToken(attr map[string]any) (string, error) {
	claims := &devJWTClaims{
		Attrs: attr,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	res, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		return "", err
	}

	return res, nil
}
