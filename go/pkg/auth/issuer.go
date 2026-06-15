package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Issuer mints RS256 JWTs for auth-service (mini-program and internal OIDC).
type Issuer struct {
	issuer   string
	audience string
	key      *rsa.PrivateKey
	kid      string
}

// IssuerConfig configures the internal token issuer.
type IssuerConfig struct {
	Issuer         string
	Audience       string
	PrivateKeyPEM  string
	GenerateIfEmpty bool
}

// NewIssuer loads or generates an RSA key for signing access tokens.
func NewIssuer(cfg IssuerConfig) (*Issuer, error) {
	key, kid, err := loadOrGenerateKey(cfg.PrivateKeyPEM, cfg.GenerateIfEmpty)
	if err != nil {
		return nil, err
	}
	return &Issuer{
		issuer:   cfg.Issuer,
		audience: cfg.Audience,
		key:      key,
		kid:      kid,
	}, nil
}

// AccessToken builds a signed JWT for the given identity fields.
func (i *Issuer) AccessToken(userID, tenantID string, roles []string, ttl time.Duration) (string, error) {
	if i == nil || i.key == nil {
		return "", fmt.Errorf("issuer not configured")
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":       userID,
		"user_id":   userID,
		"tenant_id": tenantID,
		"roles":     roles,
		"iat":       now.Unix(),
		"exp":       now.Add(ttl).Unix(),
	}
	if i.issuer != "" {
		claims["iss"] = i.issuer
	}
	if i.audience != "" {
		claims["aud"] = i.audience
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = i.kid
	return tok.SignedString(i.key)
}

// JWKSJSON returns a JWKS document for Gateway verification.
func (i *Issuer) JWKSJSON() ([]byte, error) {
	if i == nil || i.key == nil {
		return nil, fmt.Errorf("issuer not configured")
	}
	n := i.key.PublicKey.N
	set := jwksSet{
		Keys: []jwkKey{{
			Kty: "RSA",
			Use: "sig",
			Alg: "RS256",
			Kid: i.kid,
			N:   base64.RawURLEncoding.EncodeToString(n.Bytes()),
			E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(i.key.PublicKey.E)).Bytes()),
		}},
	}
	return json.Marshal(set)
}

type jwksSet struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func loadOrGenerateKey(pemStr string, allowGenerate bool) (*rsa.PrivateKey, string, error) {
	if pemStr != "" {
		block, _ := pem.Decode([]byte(pemStr))
		if block == nil {
			return nil, "", fmt.Errorf("invalid AUTH_JWT_PRIVATE_KEY_PEM")
		}
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			pk, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err2 != nil {
				return nil, "", fmt.Errorf("parse private key: %w", err)
			}
			rsaKey, ok := pk.(*rsa.PrivateKey)
			if !ok {
				return nil, "", fmt.Errorf("private key is not RSA")
			}
			key = rsaKey
		}
		return key, "auth-rsa-1", nil
	}
	if !allowGenerate {
		return nil, "", fmt.Errorf("AUTH_JWT_PRIVATE_KEY_PEM required in production")
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, "", err
	}
	return key, "auth-dev-rsa", nil
}
