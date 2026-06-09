package middleware

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	errInvalidTokenFormat  = errors.New("invalid JWT format: expected 3 dot-separated parts")
	errUnsupportedAlg      = errors.New("unsupported signing algorithm: only RS256 is supported")
	errMissingKID          = errors.New("JWT header missing kid (key ID)")
	errKeyNotFound         = errors.New("key ID not found in JWKS")
	errInvalidSignature    = errors.New("JWT signature verification failed")
	errTokenExpired        = errors.New("token has expired")
	errJWKSFetchFailed     = errors.New("failed to fetch JWKS")
	errInvalidJWKSResponse = errors.New("invalid JWKS response")
	errInvalidRSAKey       = errors.New("invalid RSA public key in JWKS")
	errInvalidIssuer       = errors.New("token issuer does not match expected issuer")
	errInvalidAudience     = errors.New("token audience does not match expected audience")
)

// jwksResponse represents the JSON Web Key Set response from the OIDC provider.
type jwksResponse struct {
	Keys []jwkKey `json:"keys"`
}

// jwkKey represents a single JSON Web Key.
type jwkKey struct {
	KTY string `json:"kty"`
	Use string `json:"use"`
	KID string `json:"kid"`
	ALG string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// jwtHeader represents the decoded JWT header.
type jwtHeader struct {
	ALG string `json:"alg"`
	KID string `json:"kid"`
	TYP string `json:"typ"`
}

// jwtPayload represents the decoded JWT payload with standard claims.
type jwtPayload struct {
	Sub string `json:"sub"`
	Exp int64  `json:"exp"`
	Iat int64  `json:"iat"`
	Iss string `json:"iss"`
	Aud any    `json:"aud"`

	// Authentik (the canonical BV IdP) models roles as GROUPS named
	// "<system>:<role>" (e.g. "analysis-bi:analyst", "social-care:admin")
	// plus the global "superadmin", delivered in the "groups" claim
	// (svc-people-context ADR-029). "roles" is kept as a generic fallback for
	// OIDC providers that expose a flat roles array.
	Groups     []string `json:"groups,omitempty"`
	RolesClaim []string `json:"roles,omitempty"`
}

// JWKSValidator validates JWT tokens against a JWKS endpoint.
// It caches the key set and refreshes periodically.
type JWKSValidator struct {
	jwksURL          string
	httpClient       *http.Client
	mu               sync.RWMutex
	keys             map[string]*rsa.PublicKey
	lastFetch        time.Time
	cacheTTL         time.Duration
	nowFunc          func() time.Time // for testing
	expectedIssuer   string
	expectedAudience string
}

// JWKSValidatorOption configures optional JWKSValidator behavior.
type JWKSValidatorOption func(*JWKSValidator)

// WithHTTPClient sets a custom HTTP client for JWKS fetching.
func WithHTTPClient(client *http.Client) JWKSValidatorOption {
	return func(v *JWKSValidator) {
		v.httpClient = client
	}
}

// WithCacheTTL sets the cache duration for JWKS keys.
func WithCacheTTL(ttl time.Duration) JWKSValidatorOption {
	return func(v *JWKSValidator) {
		v.cacheTTL = ttl
	}
}

// WithIssuer sets the expected issuer (iss) claim. Tokens with a different
// issuer will be rejected. If not set, issuer validation is skipped.
func WithIssuer(iss string) JWKSValidatorOption {
	return func(v *JWKSValidator) {
		v.expectedIssuer = iss
	}
}

// WithAudience sets the expected audience (aud) claim. Tokens that do not
// contain this audience will be rejected. If not set, audience validation
// is skipped.
func WithAudience(aud string) JWKSValidatorOption {
	return func(v *JWKSValidator) {
		v.expectedAudience = aud
	}
}

// withNowFunc overrides the time source (testing only).
func withNowFunc(fn func() time.Time) JWKSValidatorOption {
	return func(v *JWKSValidator) {
		v.nowFunc = fn
	}
}

// NewJWKSValidator creates a new JWKSValidator that fetches RSA public keys
// from the given JWKS URL and caches them for 5 minutes by default.
func NewJWKSValidator(jwksURL string, opts ...JWKSValidatorOption) *JWKSValidator {
	v := &JWKSValidator{
		jwksURL:    jwksURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		keys:       make(map[string]*rsa.PublicKey),
		cacheTTL:   5 * time.Minute,
		nowFunc:    time.Now,
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// Validate parses and validates the raw JWT token string.
// It verifies the RSA-SHA256 signature against keys from the JWKS endpoint,
// checks expiration, and extracts subject and roles.
func (v *JWKSValidator) Validate(ctx context.Context, tokenString string) (*Claims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errInvalidTokenFormat
	}

	// Decode and parse header
	header, err := decodeJWTHeader(parts[0])
	if err != nil {
		return nil, fmt.Errorf("decode JWT header: %w", err)
	}

	if header.ALG != "RS256" {
		return nil, errUnsupportedAlg
	}

	if header.KID == "" {
		return nil, errMissingKID
	}

	// Look up the signing key
	pubKey, err := v.getKey(ctx, header.KID)
	if err != nil {
		return nil, err
	}

	// Verify RSA-SHA256 signature
	signedContent := parts[0] + "." + parts[1]
	signatureBytes, err := base64URLDecode(parts[2])
	if err != nil {
		return nil, fmt.Errorf("decode signature: %w", err)
	}

	hash := sha256.Sum256([]byte(signedContent))
	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], signatureBytes); err != nil {
		return nil, errInvalidSignature
	}

	// Decode and validate payload
	payload, err := decodeJWTPayload(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode JWT payload: %w", err)
	}

	// Check expiration
	now := v.nowFunc()
	if payload.Exp > 0 && now.Unix() > payload.Exp {
		return nil, errTokenExpired
	}

	// Validate issuer if configured
	if v.expectedIssuer != "" && payload.Iss != v.expectedIssuer {
		return nil, errInvalidIssuer
	}

	// Validate audience if configured
	if v.expectedAudience != "" && !audienceContains(payload.Aud, v.expectedAudience) {
		return nil, errInvalidAudience
	}

	// Extract roles from various claim formats
	roles := extractRoles(payload)

	return &Claims{
		Subject: payload.Sub,
		Roles:   roles,
	}, nil
}

// getKey looks up an RSA public key by kid, refreshing the cache if needed.
// Security: does NOT fall back to stale keys on refresh failure (fail-closed).
func (v *JWKSValidator) getKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	// Try cache first
	v.mu.RLock()
	key, found := v.keys[kid]
	expired := v.nowFunc().After(v.lastFetch.Add(v.cacheTTL))
	v.mu.RUnlock()

	if found && !expired {
		return key, nil
	}

	// Refresh keys from JWKS endpoint (fail-closed: reject if refresh fails)
	if err := v.refreshKeys(ctx); err != nil {
		return nil, err
	}

	v.mu.RLock()
	key, found = v.keys[kid]
	v.mu.RUnlock()

	if !found {
		return nil, errKeyNotFound
	}

	return key, nil
}

// refreshKeys fetches the JWKS document and updates the key cache.
func (v *JWKSValidator) refreshKeys(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", errJWKSFetchFailed, err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", errJWKSFetchFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: HTTP %d", errJWKSFetchFailed, resp.StatusCode)
	}

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("%w: %v", errInvalidJWKSResponse, err)
	}

	keys := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, jwk := range jwks.Keys {
		if jwk.KTY != "RSA" {
			continue
		}
		if jwk.Use != "" && jwk.Use != "sig" {
			continue
		}

		pubKey, err := parseRSAPublicKey(jwk)
		if err != nil {
			continue // skip invalid keys rather than failing entirely
		}
		keys[jwk.KID] = pubKey
	}

	v.mu.Lock()
	v.keys = keys
	v.lastFetch = v.nowFunc()
	v.mu.Unlock()

	return nil
}

// parseRSAPublicKey converts a JWK to an *rsa.PublicKey.
func parseRSAPublicKey(jwk jwkKey) (*rsa.PublicKey, error) {
	nBytes, err := base64URLDecode(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid modulus: %v", errInvalidRSAKey, err)
	}

	eBytes, err := base64URLDecode(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid exponent: %v", errInvalidRSAKey, err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	if !e.IsInt64() || e.Int64() > 1<<31-1 {
		return nil, fmt.Errorf("%w: exponent too large", errInvalidRSAKey)
	}

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}

// decodeJWTHeader decodes the base64url-encoded JWT header.
func decodeJWTHeader(encoded string) (jwtHeader, error) {
	data, err := base64URLDecode(encoded)
	if err != nil {
		return jwtHeader{}, err
	}
	var h jwtHeader
	if err := json.Unmarshal(data, &h); err != nil {
		return jwtHeader{}, err
	}
	return h, nil
}

// decodeJWTPayload decodes the base64url-encoded JWT payload.
func decodeJWTPayload(encoded string) (jwtPayload, error) {
	data, err := base64URLDecode(encoded)
	if err != nil {
		return jwtPayload{}, err
	}
	var p jwtPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return jwtPayload{}, err
	}
	return p, nil
}

// base64URLDecode decodes a base64url-encoded string (without padding).
func base64URLDecode(s string) ([]byte, error) {
	// Add padding if necessary
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}

// extractRoles extracts roles from the validated JWT payload. Authentik
// "groups" is the canonical BV claim; a flat "roles" array is accepted as a
// generic fallback. Returns nil when no role claim is present.
func extractRoles(p jwtPayload) []string {
	// Authentik (canonical BV IdP): groups "<system>:<role>" (e.g.
	// "analysis-bi:analyst") + "superadmin" (people-context ADR-029).
	if len(p.Groups) > 0 {
		return p.Groups
	}

	// Generic fallback: flat "roles" array.
	if len(p.RolesClaim) > 0 {
		return p.RolesClaim
	}

	return nil
}

// audienceContains checks whether the JWT aud claim (which can be a string
// or an array of strings) contains the expected audience value.
func audienceContains(aud any, expected string) bool {
	switch v := aud.(type) {
	case string:
		return v == expected
	case []any:
		for _, a := range v {
			if s, ok := a.(string); ok && s == expected {
				return true
			}
		}
	case []string:
		for _, a := range v {
			if a == expected {
				return true
			}
		}
	}
	return false
}
