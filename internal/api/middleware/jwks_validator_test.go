package middleware

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// testKeyPair holds an RSA key pair for testing.
type testKeyPair struct {
	Private *rsa.PrivateKey
	Public  *rsa.PublicKey
	KID     string
}

func generateTestKey(t *testing.T, kid string) testKeyPair {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	return testKeyPair{
		Private: priv,
		Public:  &priv.PublicKey,
		KID:     kid,
	}
}

func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// signJWT creates a signed JWT token using the given key pair and payload.
func signJWT(t *testing.T, key testKeyPair, headerOverrides map[string]any, payload map[string]any) string {
	t.Helper()

	header := map[string]any{
		"alg": "RS256",
		"typ": "JWT",
		"kid": key.KID,
	}
	for k, v := range headerOverrides {
		header[k] = v
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("marshal header: %v", err)
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	headerB64 := base64URLEncode(headerJSON)
	payloadB64 := base64URLEncode(payloadJSON)
	signedContent := headerB64 + "." + payloadB64

	hash := sha256.Sum256([]byte(signedContent))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key.Private, crypto.SHA256, hash[:])
	if err != nil {
		t.Fatalf("sign JWT: %v", err)
	}

	return signedContent + "." + base64URLEncode(sig)
}

// serveJWKS starts an httptest.Server that serves a JWKS document
// containing the given RSA public keys.
func serveJWKS(t *testing.T, keys ...testKeyPair) *httptest.Server {
	t.Helper()

	type jwkKeyJSON struct {
		KTY string `json:"kty"`
		Use string `json:"use"`
		KID string `json:"kid"`
		ALG string `json:"alg"`
		N   string `json:"n"`
		E   string `json:"e"`
	}

	var jwkKeys []jwkKeyJSON
	for _, k := range keys {
		eBytes := big.NewInt(int64(k.Public.E)).Bytes()
		jwkKeys = append(jwkKeys, jwkKeyJSON{
			KTY: "RSA",
			Use: "sig",
			KID: k.KID,
			ALG: "RS256",
			N:   base64URLEncode(k.Public.N.Bytes()),
			E:   base64URLEncode(eBytes),
		})
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": jwkKeys})
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestJWKSValidator_ValidToken(t *testing.T) {
	key := generateTestKey(t, "test-key-1")
	jwksSrv := serveJWKS(t, key)

	now := time.Now()
	validator := NewJWKSValidator(jwksSrv.URL,
		withNowFunc(func() time.Time { return now }),
	)

	token := signJWT(t, key, nil, map[string]any{
		"sub":   "user-123",
		"exp":   now.Add(1 * time.Hour).Unix(),
		"iat":   now.Unix(),
		"roles": []string{"analyst", "admin"},
	})

	claims, err := validator.Validate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if claims.Subject != "user-123" {
		t.Errorf("subject = %q, want %q", claims.Subject, "user-123")
	}

	if len(claims.Roles) != 2 || claims.Roles[0] != "analyst" || claims.Roles[1] != "admin" {
		t.Errorf("roles = %v, want [analyst admin]", claims.Roles)
	}
}

func TestJWKSValidator_InvalidTokenFormat(t *testing.T) {
	validator := NewJWKSValidator("http://unused")

	tests := []struct {
		name  string
		token string
	}{
		{"empty string", ""},
		{"single part", "abc"},
		{"two parts", "abc.def"},
		{"four parts", "a.b.c.d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.Validate(context.Background(), tt.token)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err != errInvalidTokenFormat {
				t.Errorf("error = %v, want errInvalidTokenFormat", err)
			}
		})
	}
}

func TestJWKSValidator_ExpiredToken(t *testing.T) {
	key := generateTestKey(t, "test-key-2")
	jwksSrv := serveJWKS(t, key)

	now := time.Now()
	validator := NewJWKSValidator(jwksSrv.URL,
		withNowFunc(func() time.Time { return now }),
	)

	// Token expired 1 hour ago
	token := signJWT(t, key, nil, map[string]any{
		"sub": "user-expired",
		"exp": now.Add(-1 * time.Hour).Unix(),
		"iat": now.Add(-2 * time.Hour).Unix(),
	})

	_, err := validator.Validate(context.Background(), token)
	if err != errTokenExpired {
		t.Errorf("error = %v, want errTokenExpired", err)
	}
}

func TestJWKSValidator_UnsupportedAlgorithm(t *testing.T) {
	key := generateTestKey(t, "test-key-3")
	jwksSrv := serveJWKS(t, key)

	validator := NewJWKSValidator(jwksSrv.URL)

	// Force a different alg in the header
	token := signJWT(t, key, map[string]any{"alg": "HS256"}, map[string]any{
		"sub": "user-hs256",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})

	_, err := validator.Validate(context.Background(), token)
	if err != errUnsupportedAlg {
		t.Errorf("error = %v, want errUnsupportedAlg", err)
	}
}

func TestJWKSValidator_MissingKID(t *testing.T) {
	key := generateTestKey(t, "test-key-4")
	jwksSrv := serveJWKS(t, key)

	validator := NewJWKSValidator(jwksSrv.URL)

	// Remove kid from header
	token := signJWT(t, key, map[string]any{"kid": ""}, map[string]any{
		"sub": "user-no-kid",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})

	_, err := validator.Validate(context.Background(), token)
	if err != errMissingKID {
		t.Errorf("error = %v, want errMissingKID", err)
	}
}

func TestJWKSValidator_KeyNotFound(t *testing.T) {
	key := generateTestKey(t, "known-key")
	jwksSrv := serveJWKS(t, key)

	now := time.Now()
	validator := NewJWKSValidator(jwksSrv.URL,
		withNowFunc(func() time.Time { return now }),
	)

	// Sign with a different kid than what the JWKS serves
	unknownKey := generateTestKey(t, "unknown-key")
	token := signJWT(t, unknownKey, nil, map[string]any{
		"sub": "user-unknown",
		"exp": now.Add(1 * time.Hour).Unix(),
	})

	_, err := validator.Validate(context.Background(), token)
	if err != errKeyNotFound {
		t.Errorf("error = %v, want errKeyNotFound", err)
	}
}

func TestJWKSValidator_InvalidSignature(t *testing.T) {
	key := generateTestKey(t, "test-key-sig")
	differentKey := generateTestKey(t, "test-key-sig")

	// Serve the public key from 'key', but sign with 'differentKey'
	jwksSrv := serveJWKS(t, key)

	now := time.Now()
	validator := NewJWKSValidator(jwksSrv.URL,
		withNowFunc(func() time.Time { return now }),
	)

	// Sign with a different private key but use the same kid
	differentKey.KID = key.KID
	token := signJWT(t, differentKey, nil, map[string]any{
		"sub": "user-badsig",
		"exp": now.Add(1 * time.Hour).Unix(),
	})

	_, err := validator.Validate(context.Background(), token)
	if err != errInvalidSignature {
		t.Errorf("error = %v, want errInvalidSignature", err)
	}
}

func TestJWKSValidator_CacheRefresh(t *testing.T) {
	key := generateTestKey(t, "cache-key")
	fetchCount := 0

	eBytes := big.NewInt(int64(key.Public.E)).Bytes()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetchCount++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]any{
				{
					"kty": "RSA",
					"use": "sig",
					"kid": key.KID,
					"alg": "RS256",
					"n":   base64URLEncode(key.Public.N.Bytes()),
					"e":   base64URLEncode(eBytes),
				},
			},
		})
	}))
	t.Cleanup(srv.Close)

	now := time.Now()
	currentTime := now
	validator := NewJWKSValidator(srv.URL,
		WithCacheTTL(1*time.Minute),
		withNowFunc(func() time.Time { return currentTime }),
	)

	makeToken := func() string {
		return signJWT(t, key, nil, map[string]any{
			"sub": "user-cache",
			"exp": now.Add(1 * time.Hour).Unix(),
		})
	}

	// First call fetches keys
	if _, err := validator.Validate(context.Background(), makeToken()); err != nil {
		t.Fatalf("first validate: %v", err)
	}
	if fetchCount != 1 {
		t.Fatalf("fetchCount after first call = %d, want 1", fetchCount)
	}

	// Second call within TTL uses cache
	if _, err := validator.Validate(context.Background(), makeToken()); err != nil {
		t.Fatalf("second validate: %v", err)
	}
	if fetchCount != 1 {
		t.Fatalf("fetchCount after second call = %d, want 1", fetchCount)
	}

	// Advance time past TTL
	currentTime = now.Add(2 * time.Minute)

	if _, err := validator.Validate(context.Background(), makeToken()); err != nil {
		t.Fatalf("third validate: %v", err)
	}
	if fetchCount != 2 {
		t.Fatalf("fetchCount after TTL expiry = %d, want 2", fetchCount)
	}
}

func TestJWKSValidator_JWKSFetchFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	validator := NewJWKSValidator(srv.URL)

	// Create a structurally valid JWT to get past header parsing
	header := base64URLEncode([]byte(`{"alg":"RS256","kid":"missing","typ":"JWT"}`))
	payload := base64URLEncode([]byte(`{"sub":"test"}`))
	token := header + "." + payload + "." + base64URLEncode([]byte("fake-sig"))

	_, err := validator.Validate(context.Background(), token)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestJWKSValidator_ZitadelRoles(t *testing.T) {
	key := generateTestKey(t, "zitadel-key")
	jwksSrv := serveJWKS(t, key)

	now := time.Now()
	validator := NewJWKSValidator(jwksSrv.URL,
		withNowFunc(func() time.Time { return now }),
	)

	token := signJWT(t, key, nil, map[string]any{
		"sub": "zitadel-user",
		"exp": now.Add(1 * time.Hour).Unix(),
		"urn:zitadel:iam:org:project:roles": []map[string]any{
			{"analyst": map[string]any{"org_id": "123"}},
			{"viewer": map[string]any{"org_id": "456"}},
		},
	})

	claims, err := validator.Validate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if claims.Subject != "zitadel-user" {
		t.Errorf("subject = %q, want %q", claims.Subject, "zitadel-user")
	}

	if len(claims.Roles) != 2 {
		t.Errorf("roles count = %d, want 2; roles = %v", len(claims.Roles), claims.Roles)
	}

	// Roles from Zitadel maps -- check both are present (order may vary)
	roleSet := make(map[string]bool)
	for _, r := range claims.Roles {
		roleSet[r] = true
	}
	if !roleSet["analyst"] || !roleSet["viewer"] {
		t.Errorf("expected roles analyst and viewer, got %v", claims.Roles)
	}
}

func TestJWKSValidator_NoRoles(t *testing.T) {
	key := generateTestKey(t, "noroles-key")
	jwksSrv := serveJWKS(t, key)

	now := time.Now()
	validator := NewJWKSValidator(jwksSrv.URL,
		withNowFunc(func() time.Time { return now }),
	)

	token := signJWT(t, key, nil, map[string]any{
		"sub": "user-noroles",
		"exp": now.Add(1 * time.Hour).Unix(),
	})

	claims, err := validator.Validate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if claims.Subject != "user-noroles" {
		t.Errorf("subject = %q, want %q", claims.Subject, "user-noroles")
	}

	if claims.Roles != nil {
		t.Errorf("roles = %v, want nil", claims.Roles)
	}
}

func TestJWKSValidator_WrongIssuer(t *testing.T) {
	key := generateTestKey(t, "iss-key")
	jwksSrv := serveJWKS(t, key)

	now := time.Now()
	validator := NewJWKSValidator(jwksSrv.URL,
		withNowFunc(func() time.Time { return now }),
		WithIssuer("https://auth.expected.com"),
	)

	token := signJWT(t, key, nil, map[string]any{
		"sub": "user-wrong-iss",
		"exp": now.Add(1 * time.Hour).Unix(),
		"iss": "https://auth.evil.com",
	})

	_, err := validator.Validate(context.Background(), token)
	if err != errInvalidIssuer {
		t.Errorf("error = %v, want errInvalidIssuer", err)
	}
}

func TestJWKSValidator_WrongAudience(t *testing.T) {
	key := generateTestKey(t, "aud-key")
	jwksSrv := serveJWKS(t, key)

	now := time.Now()
	validator := NewJWKSValidator(jwksSrv.URL,
		withNowFunc(func() time.Time { return now }),
		WithAudience("my-api"),
	)

	token := signJWT(t, key, nil, map[string]any{
		"sub": "user-wrong-aud",
		"exp": now.Add(1 * time.Hour).Unix(),
		"aud": "other-api",
	})

	_, err := validator.Validate(context.Background(), token)
	if err != errInvalidAudience {
		t.Errorf("error = %v, want errInvalidAudience", err)
	}
}

func TestJWKSValidator_WrongAudienceArray(t *testing.T) {
	key := generateTestKey(t, "aud-arr-key")
	jwksSrv := serveJWKS(t, key)

	now := time.Now()
	validator := NewJWKSValidator(jwksSrv.URL,
		withNowFunc(func() time.Time { return now }),
		WithAudience("my-api"),
	)

	token := signJWT(t, key, nil, map[string]any{
		"sub": "user-wrong-aud-arr",
		"exp": now.Add(1 * time.Hour).Unix(),
		"aud": []string{"other-api", "another-api"},
	})

	_, err := validator.Validate(context.Background(), token)
	if err != errInvalidAudience {
		t.Errorf("error = %v, want errInvalidAudience", err)
	}
}

func TestJWKSValidator_CorrectIssuerAndAudience(t *testing.T) {
	key := generateTestKey(t, "issaud-key")
	jwksSrv := serveJWKS(t, key)

	now := time.Now()
	validator := NewJWKSValidator(jwksSrv.URL,
		withNowFunc(func() time.Time { return now }),
		WithIssuer("https://auth.acdgbrasil.com.br"),
		WithAudience("analysis-bi"),
	)

	token := signJWT(t, key, nil, map[string]any{
		"sub":   "user-correct",
		"exp":   now.Add(1 * time.Hour).Unix(),
		"iss":   "https://auth.acdgbrasil.com.br",
		"aud":   []string{"analysis-bi", "other-service"},
		"roles": []string{"analyst"},
	})

	claims, err := validator.Validate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if claims.Subject != "user-correct" {
		t.Errorf("subject = %q, want %q", claims.Subject, "user-correct")
	}
}

func TestJWKSValidator_AudienceAsString(t *testing.T) {
	key := generateTestKey(t, "aud-str-key")
	jwksSrv := serveJWKS(t, key)

	now := time.Now()
	validator := NewJWKSValidator(jwksSrv.URL,
		withNowFunc(func() time.Time { return now }),
		WithAudience("analysis-bi"),
	)

	token := signJWT(t, key, nil, map[string]any{
		"sub": "user-aud-str",
		"exp": now.Add(1 * time.Hour).Unix(),
		"aud": "analysis-bi",
	})

	claims, err := validator.Validate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if claims.Subject != "user-aud-str" {
		t.Errorf("subject = %q, want %q", claims.Subject, "user-aud-str")
	}
}

func TestJWKSValidator_NoIssuerConfigured(t *testing.T) {
	key := generateTestKey(t, "noiss-key")
	jwksSrv := serveJWKS(t, key)

	now := time.Now()
	// No WithIssuer — issuer validation skipped
	validator := NewJWKSValidator(jwksSrv.URL,
		withNowFunc(func() time.Time { return now }),
	)

	token := signJWT(t, key, nil, map[string]any{
		"sub": "user-anyiss",
		"exp": now.Add(1 * time.Hour).Unix(),
		"iss": "https://any-issuer.com",
	})

	_, err := validator.Validate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error when issuer not configured: %v", err)
	}
}

func TestDecodeJWTHeader(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantALG string
		wantKID string
		wantErr bool
	}{
		{
			name:    "valid header",
			input:   base64URLEncode([]byte(`{"alg":"RS256","kid":"key-1","typ":"JWT"}`)),
			wantALG: "RS256",
			wantKID: "key-1",
		},
		{
			name:    "invalid base64",
			input:   "!!!not-base64!!!",
			wantErr: true,
		},
		{
			name:    "invalid json",
			input:   base64URLEncode([]byte(`not json`)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := decodeJWTHeader(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if h.ALG != tt.wantALG {
				t.Errorf("alg = %q, want %q", h.ALG, tt.wantALG)
			}
			if h.KID != tt.wantKID {
				t.Errorf("kid = %q, want %q", h.KID, tt.wantKID)
			}
		})
	}
}

func TestParseRSAPublicKey(t *testing.T) {
	key := generateTestKey(t, "parse-key")
	eBytes := big.NewInt(int64(key.Public.E)).Bytes()

	jwk := jwkKey{
		KTY: "RSA",
		KID: "parse-key",
		N:   base64URLEncode(key.Public.N.Bytes()),
		E:   base64URLEncode(eBytes),
	}

	pubKey, err := parseRSAPublicKey(jwk)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pubKey.N.Cmp(key.Public.N) != 0 {
		t.Error("modulus mismatch")
	}
	if pubKey.E != key.Public.E {
		t.Errorf("exponent = %d, want %d", pubKey.E, key.Public.E)
	}
}

func TestParseRSAPublicKey_InvalidModulus(t *testing.T) {
	jwk := jwkKey{
		KTY: "RSA",
		KID: "bad-n",
		N:   "!!!invalid!!!",
		E:   base64URLEncode([]byte{1, 0, 1}),
	}

	_, err := parseRSAPublicKey(jwk)
	if err == nil {
		t.Fatal("expected error for invalid modulus")
	}
}

func TestExtractRoles(t *testing.T) {
	tests := []struct {
		name     string
		payload  jwtPayload
		expected []string
	}{
		{
			name:     "direct roles claim",
			payload:  jwtPayload{RolesClaim: []string{"admin", "viewer"}},
			expected: []string{"admin", "viewer"},
		},
		{
			name: "zitadel roles",
			payload: jwtPayload{
				ZitadelRoles: []map[string]any{
					{"editor": map[string]any{}},
				},
			},
			expected: []string{"editor"},
		},
		{
			name: "keycloak realm_access",
			payload: jwtPayload{
				RealmAccess: &realmAccess{Roles: []string{"realm-admin"}},
			},
			expected: []string{"realm-admin"},
		},
		{
			name:     "no roles",
			payload:  jwtPayload{},
			expected: nil,
		},
		{
			name: "direct takes priority over zitadel",
			payload: jwtPayload{
				RolesClaim:   []string{"direct-role"},
				ZitadelRoles: []map[string]any{{"zitadel-role": map[string]any{}}},
			},
			expected: []string{"direct-role"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roles := extractRoles(tt.payload)
			if fmt.Sprintf("%v", roles) != fmt.Sprintf("%v", tt.expected) {
				t.Errorf("roles = %v, want %v", roles, tt.expected)
			}
		})
	}
}
