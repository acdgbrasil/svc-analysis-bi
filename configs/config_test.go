package configs

import (
	"net/url"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Load() tests
// ---------------------------------------------------------------------------

func TestLoad_DefaultsAppliedWithSalt(t *testing.T) {
	// When only PATIENT_HASH_SALT is set, all other fields use defaults.
	t.Setenv("PATIENT_HASH_SALT", "test-salt-value")
	// Clear any potentially-set env vars that would override defaults.
	t.Setenv("PORT", "")
	t.Setenv("SERVER_HOST", "")
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_NAME", "")
	t.Setenv("DB_MAX_CONNS", "")
	t.Setenv("DB_IDLE_TIMEOUT", "")
	t.Setenv("NATS_URL", "")
	t.Setenv("NATS_STREAM", "")
	t.Setenv("NATS_CONSUMER", "")
	t.Setenv("LOG_LEVEL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Server.Port default", cfg.Server.Port, 8080},
		{"Server.Host default", cfg.Server.Host, "0.0.0.0"},
		{"Database.Host default", cfg.Database.Host, "localhost"},
		{"Database.Port default", cfg.Database.Port, 5432},
		{"Database.Name default", cfg.Database.Name, "analysis_bi"},
		{"Database.MaxConns default", cfg.Database.MaxConns, 10},
		{"NATS.URL default", cfg.NATS.URL, "nats://localhost:4222"},
		{"NATS.Stream default", cfg.NATS.Stream, "SOCIAL_CARE_EVENTS"},
		{"NATS.Consumer default", cfg.NATS.Consumer, "analysis-bi"},
		{"PatientHashSalt set", cfg.PatientHashSalt, "test-salt-value"},
		{"LogLevel default", cfg.LogLevel, "info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestLoad_DefaultIdleTimeout(t *testing.T) {
	t.Setenv("PATIENT_HASH_SALT", "salt")
	t.Setenv("DB_IDLE_TIMEOUT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	// Default is 5 minutes per contract.
	wantMinutes := 5.0
	gotMinutes := cfg.Database.IdleTimeout.Minutes()
	if gotMinutes != wantMinutes {
		t.Errorf("Database.IdleTimeout = %v, want %v minutes", cfg.Database.IdleTimeout, wantMinutes)
	}
}

func TestLoad_CustomEnvVarsOverrideDefaults(t *testing.T) {
	t.Setenv("PATIENT_HASH_SALT", "custom-salt")
	t.Setenv("PORT", "9090")
	t.Setenv("SERVER_HOST", "127.0.0.1")
	t.Setenv("DB_HOST", "db.example.com")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "myuser")
	t.Setenv("DB_PASSWORD", "mypass")
	t.Setenv("DB_NAME", "custom_db")
	t.Setenv("DB_MAX_CONNS", "20")
	t.Setenv("DB_IDLE_TIMEOUT", "10m")
	t.Setenv("NATS_URL", "nats://nats.example.com:4222")
	t.Setenv("NATS_STREAM", "CUSTOM_STREAM")
	t.Setenv("NATS_CONSUMER", "custom-consumer")
	t.Setenv("JWKS_URL", "https://auth.example.com/keys")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Server.Port", cfg.Server.Port, 9090},
		{"Server.Host", cfg.Server.Host, "127.0.0.1"},
		{"Database.Host", cfg.Database.Host, "db.example.com"},
		{"Database.Port", cfg.Database.Port, 5433},
		{"Database.User", cfg.Database.User, "myuser"},
		{"Database.Password", cfg.Database.Password, "mypass"},
		{"Database.Name", cfg.Database.Name, "custom_db"},
		{"Database.MaxConns", cfg.Database.MaxConns, 20},
		{"NATS.URL", cfg.NATS.URL, "nats://nats.example.com:4222"},
		{"NATS.Stream", cfg.NATS.Stream, "CUSTOM_STREAM"},
		{"NATS.Consumer", cfg.NATS.Consumer, "custom-consumer"},
		{"Auth.JWKSUrl", cfg.Auth.JWKSUrl, "https://auth.example.com/keys"},
		{"PatientHashSalt", cfg.PatientHashSalt, "custom-salt"},
		{"LogLevel", cfg.LogLevel, "debug"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}

	// Separately check IdleTimeout because it is a time.Duration.
	wantMinutes := 10.0
	gotMinutes := cfg.Database.IdleTimeout.Minutes()
	if gotMinutes != wantMinutes {
		t.Errorf("Database.IdleTimeout = %v, want %v minutes", cfg.Database.IdleTimeout, wantMinutes)
	}
}

func TestLoad_MissingSalt(t *testing.T) {
	t.Setenv("PATIENT_HASH_SALT", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for missing PATIENT_HASH_SALT, got nil")
	}
	if !strings.Contains(err.Error(), "PATIENT_HASH_SALT") {
		t.Errorf("Load() error = %q, want mention of PATIENT_HASH_SALT", err)
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	tests := []struct {
		name    string
		portVal string
	}{
		{"port zero", "0"},
		{"port above 65535", "65536"},
		{"port negative", "-1"},
		{"port non-numeric", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PATIENT_HASH_SALT", "salt")
			t.Setenv("PORT", tt.portVal)

			_, err := Load()
			if err == nil {
				t.Fatalf("Load() expected error for PORT=%q, got nil", tt.portVal)
			}
		})
	}
}

func TestLoad_InvalidMaxConns(t *testing.T) {
	t.Setenv("PATIENT_HASH_SALT", "salt")
	t.Setenv("DB_MAX_CONNS", "0")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for DB_MAX_CONNS=0, got nil")
	}
}

func TestLoad_NegativeMaxConns(t *testing.T) {
	t.Setenv("PATIENT_HASH_SALT", "salt")
	t.Setenv("DB_MAX_CONNS", "-5")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for DB_MAX_CONNS=-5, got nil")
	}
}

func TestLoad_InvalidIdleTimeout(t *testing.T) {
	t.Setenv("PATIENT_HASH_SALT", "salt")
	t.Setenv("DB_IDLE_TIMEOUT", "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for DB_IDLE_TIMEOUT=invalid, got nil")
	}
}

// ---------------------------------------------------------------------------
// Validate() tests
// ---------------------------------------------------------------------------

func TestValidate_ValidConfig(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{Port: 8080, Host: "0.0.0.0"},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "user",
			Password: "pass",
			Name:     "analysis_bi",
			MaxConns: 10,
		},
		NATS:            NATSConfig{URL: "nats://localhost:4222", Stream: "S", Consumer: "C"},
		PatientHashSalt: "salt",
		LogLevel:        "info",
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}

func TestValidate_EmptyPatientHashSalt(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{Port: 8080, Host: "0.0.0.0"},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "user",
			Password: "pass",
			Name:     "analysis_bi",
			MaxConns: 10,
		},
		PatientHashSalt: "",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() expected ErrMissingSalt, got nil")
	}
	if err != ErrMissingSalt {
		t.Errorf("Validate() error = %v, want %v", err, ErrMissingSalt)
	}
}

func TestValidate_PortZero(t *testing.T) {
	cfg := Config{
		Server:          ServerConfig{Port: 0, Host: "0.0.0.0"},
		Database:        DatabaseConfig{User: "u", Password: "p", MaxConns: 1},
		PatientHashSalt: "salt",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() expected ErrInvalidPort, got nil")
	}
	if err != ErrInvalidPort {
		t.Errorf("Validate() error = %v, want %v", err, ErrInvalidPort)
	}
}

func TestValidate_PortAboveMax(t *testing.T) {
	cfg := Config{
		Server:          ServerConfig{Port: 65536, Host: "0.0.0.0"},
		Database:        DatabaseConfig{User: "u", Password: "p", MaxConns: 1},
		PatientHashSalt: "salt",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() expected ErrInvalidPort, got nil")
	}
	if err != ErrInvalidPort {
		t.Errorf("Validate() error = %v, want %v", err, ErrInvalidPort)
	}
}

func TestValidate_MissingDBUser(t *testing.T) {
	cfg := Config{
		Server:          ServerConfig{Port: 8080, Host: "0.0.0.0"},
		Database:        DatabaseConfig{User: "", Password: "pass", MaxConns: 1},
		PatientHashSalt: "salt",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() expected ErrInvalidDBConfig, got nil")
	}
	if err != ErrInvalidDBConfig {
		t.Errorf("Validate() error = %v, want %v", err, ErrInvalidDBConfig)
	}
}

func TestValidate_MissingDBPassword(t *testing.T) {
	cfg := Config{
		Server:          ServerConfig{Port: 8080, Host: "0.0.0.0"},
		Database:        DatabaseConfig{User: "user", Password: "", MaxConns: 1},
		PatientHashSalt: "salt",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() expected ErrInvalidDBConfig, got nil")
	}
	if err != ErrInvalidDBConfig {
		t.Errorf("Validate() error = %v, want %v", err, ErrInvalidDBConfig)
	}
}

func TestValidate_InvalidMaxConns(t *testing.T) {
	cfg := Config{
		Server:          ServerConfig{Port: 8080, Host: "0.0.0.0"},
		Database:        DatabaseConfig{User: "u", Password: "p", MaxConns: 0},
		PatientHashSalt: "salt",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() expected ErrInvalidMaxConns, got nil")
	}
	if err != ErrInvalidMaxConns {
		t.Errorf("Validate() error = %v, want %v", err, ErrInvalidMaxConns)
	}
}

// ---------------------------------------------------------------------------
// DSN() tests
// ---------------------------------------------------------------------------

func TestLoad_AuthRequiredDefaultTrue(t *testing.T) {
	t.Setenv("PATIENT_HASH_SALT", "salt")
	t.Setenv("AUTH_REQUIRED", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if !cfg.Auth.AuthRequired {
		t.Error("Auth.AuthRequired should default to true")
	}
}

func TestValidate_AuthRequiredWithoutJWKS_Fails(t *testing.T) {
	cfg := Config{
		PatientHashSalt: "salt",
		Server:          ServerConfig{Port: 8080},
		Database:        DatabaseConfig{User: "u", Password: "p", MaxConns: 5},
		Auth:            AuthConfig{AuthRequired: true, JWKSUrl: ""},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() should fail when AUTH_REQUIRED=true and JWKS_URL is empty")
	}
	if err != ErrMissingJWKS {
		t.Errorf("Validate() error = %v, want %v", err, ErrMissingJWKS)
	}
}

func TestValidate_AuthRequiredWithJWKS_Passes(t *testing.T) {
	cfg := Config{
		PatientHashSalt: "salt",
		Server:          ServerConfig{Port: 8080},
		Database:        DatabaseConfig{User: "u", Password: "p", MaxConns: 5},
		Auth:            AuthConfig{AuthRequired: true, JWKSUrl: "https://auth.example.com/keys"},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}
}

func TestLoad_AuthRequiredFalseWithNoJWKS(t *testing.T) {
	t.Setenv("PATIENT_HASH_SALT", "salt")
	t.Setenv("AUTH_REQUIRED", "false")
	t.Setenv("JWKS_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.Auth.AuthRequired {
		t.Error("Auth.AuthRequired should be false when AUTH_REQUIRED=false")
	}
	if cfg.Auth.JWKSUrl != "" {
		t.Errorf("Auth.JWKSUrl = %q, want empty", cfg.Auth.JWKSUrl)
	}
}

func TestLoad_AuthIssuerAndAudience(t *testing.T) {
	t.Setenv("PATIENT_HASH_SALT", "salt")
	t.Setenv("AUTH_ISSUER", "https://auth.example.com")
	t.Setenv("AUTH_AUDIENCE", "my-api")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.Auth.ExpectedIssuer != "https://auth.example.com" {
		t.Errorf("Auth.ExpectedIssuer = %q, want %q", cfg.Auth.ExpectedIssuer, "https://auth.example.com")
	}
	if cfg.Auth.ExpectedAudience != "my-api" {
		t.Errorf("Auth.ExpectedAudience = %q, want %q", cfg.Auth.ExpectedAudience, "my-api")
	}
}

// ---------------------------------------------------------------------------
// DSN() tests
// ---------------------------------------------------------------------------

func TestDSN_BasicFormat(t *testing.T) {
	db := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "myuser",
		Password: "mypassword",
		Name:     "mydb",
	}

	got := db.DSN()
	want := "postgres://myuser:mypassword@localhost:5432/mydb"
	if got != want {
		t.Errorf("DSN() = %q, want %q", got, want)
	}
}

func TestDSN_SpecialCharsInPassword(t *testing.T) {
	db := DatabaseConfig{
		Host:     "db.host",
		Port:     5432,
		User:     "admin",
		Password: "p@ss:w0rd/special#",
		Name:     "testdb",
	}

	got := db.DSN()

	// The DSN should be a valid URL. Parse it and verify the password round-trips.
	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("DSN() produced unparseable URL: %v", err)
	}

	gotPassword, ok := u.User.Password()
	if !ok {
		t.Fatal("DSN() URL has no password component")
	}
	if gotPassword != db.Password {
		t.Errorf("DSN() password = %q, want %q", gotPassword, db.Password)
	}
	if u.User.Username() != db.User {
		t.Errorf("DSN() username = %q, want %q", u.User.Username(), db.User)
	}
	if u.Hostname() != db.Host {
		t.Errorf("DSN() host = %q, want %q", u.Hostname(), db.Host)
	}
	if u.Path != "/"+db.Name {
		t.Errorf("DSN() path = %q, want %q", u.Path, "/"+db.Name)
	}
}

func TestDSN_CustomPort(t *testing.T) {
	db := DatabaseConfig{
		Host:     "remote.host",
		Port:     5433,
		User:     "user",
		Password: "pass",
		Name:     "db",
	}

	got := db.DSN()
	if !strings.Contains(got, ":5433/") {
		t.Errorf("DSN() = %q, expected port 5433", got)
	}
}
