package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestHashPatientID(t *testing.T) {
	tests := []struct {
		name      string
		patientID string
		salt      string
		wantErr   error
	}{
		{
			name:      "deterministic: same id and salt produce same hash",
			patientID: "patient-123",
			salt:      "secret-salt",
			wantErr:   nil,
		},
		{
			name:      "empty salt returns ErrEmptySalt",
			patientID: "patient-123",
			salt:      "",
			wantErr:   ErrEmptySalt,
		},
		{
			name:      "empty patientID with valid salt still succeeds",
			patientID: "",
			salt:      "secret-salt",
			wantErr:   nil,
		},
		{
			name:      "whitespace-only salt is accepted (not empty)",
			patientID: "patient-123",
			salt:      "   ",
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HashPatientID(tt.patientID, tt.salt)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("HashPatientID(%q, %q) expected error %v, got nil", tt.patientID, tt.salt, tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("HashPatientID(%q, %q) error = %v, want %v", tt.patientID, tt.salt, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("HashPatientID(%q, %q) unexpected error: %v", tt.patientID, tt.salt, err)
			}
			if got == "" {
				t.Errorf("HashPatientID(%q, %q) returned empty hash", tt.patientID, tt.salt)
			}
		})
	}
}

func TestHashPatientID_Deterministic(t *testing.T) {
	hash1, err := HashPatientID("patient-abc", "my-salt")
	if err != nil {
		t.Fatalf("first call: unexpected error: %v", err)
	}
	hash2, err := HashPatientID("patient-abc", "my-salt")
	if err != nil {
		t.Fatalf("second call: unexpected error: %v", err)
	}
	if hash1 != hash2 {
		t.Errorf("same (id, salt) produced different hashes: %q vs %q", hash1, hash2)
	}
}

func TestHashPatientID_DifferentIDsDifferentHashes(t *testing.T) {
	hash1, err := HashPatientID("patient-001", "salt")
	if err != nil {
		t.Fatalf("first call: unexpected error: %v", err)
	}
	hash2, err := HashPatientID("patient-002", "salt")
	if err != nil {
		t.Fatalf("second call: unexpected error: %v", err)
	}
	if hash1 == hash2 {
		t.Errorf("different IDs produced the same hash: %q", hash1)
	}
}

func TestHashPatientID_DifferentSaltsDifferentHashes(t *testing.T) {
	hash1, err := HashPatientID("patient-001", "salt-A")
	if err != nil {
		t.Fatalf("first call: unexpected error: %v", err)
	}
	hash2, err := HashPatientID("patient-001", "salt-B")
	if err != nil {
		t.Fatalf("second call: unexpected error: %v", err)
	}
	if hash1 == hash2 {
		t.Errorf("different salts produced the same hash: %q", hash1)
	}
}

func TestHashPatientID_Is64HexChars(t *testing.T) {
	hash, err := HashPatientID("patient-xyz", "some-salt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(hash)
	if len(s) != 64 {
		t.Errorf("hash length = %d, want 64 (SHA-256 hex)", len(s))
	}
	for i, c := range s {
		if !strings.ContainsRune("0123456789abcdef", c) {
			t.Errorf("hash char at index %d is %q, not a lowercase hex digit", i, string(c))
			break
		}
	}
}

func TestGeneralizeAge(t *testing.T) {
	now := time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		birthDate time.Time
		refDate   time.Time
		wantLabel string
		wantErr   error
	}{
		{
			name:      "age 0 (born same year, same day) maps to 0-4",
			birthDate: time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC),
			refDate:   now,
			wantLabel: "0-4",
		},
		{
			name:      "age 3 maps to 0-4",
			birthDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			refDate:   now,
			wantLabel: "0-4",
		},
		{
			name:      "age 5 maps to 5-9",
			birthDate: time.Date(2021, 4, 11, 0, 0, 0, 0, time.UTC),
			refDate:   now,
			wantLabel: "5-9",
		},
		{
			name:      "age 9 maps to 5-9",
			birthDate: time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC),
			refDate:   now,
			wantLabel: "5-9",
		},
		{
			name:      "age 10 maps to 10-14",
			birthDate: time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC),
			refDate:   now,
			wantLabel: "10-14",
		},
		{
			name:      "age 79 maps to 75-79",
			birthDate: time.Date(1947, 1, 1, 0, 0, 0, 0, time.UTC),
			refDate:   now,
			wantLabel: "75-79",
		},
		{
			name:      "age 80 maps to 80+",
			birthDate: time.Date(1946, 1, 1, 0, 0, 0, 0, time.UTC),
			refDate:   now,
			wantLabel: "80+",
		},
		{
			name:      "age 95 maps to 80+",
			birthDate: time.Date(1931, 1, 1, 0, 0, 0, 0, time.UTC),
			refDate:   now,
			wantLabel: "80+",
		},
		{
			name:      "birth date in future returns ErrInvalidBirthDate",
			birthDate: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
			refDate:   now,
			wantErr:   ErrInvalidBirthDate,
		},
		{
			name:      "exactly 5 years ago today maps to 5-9",
			birthDate: time.Date(2021, 4, 11, 0, 0, 0, 0, time.UTC),
			refDate:   now,
			wantLabel: "5-9",
		},
		{
			name:      "born one day after 5-year boundary still maps to 0-4",
			birthDate: time.Date(2021, 4, 12, 0, 0, 0, 0, time.UTC),
			refDate:   now,
			wantLabel: "0-4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GeneralizeAge(tt.birthDate, tt.refDate)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("GeneralizeAge(%v, %v) expected error %v, got nil", tt.birthDate, tt.refDate, tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("GeneralizeAge(%v, %v) error = %v, want %v", tt.birthDate, tt.refDate, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("GeneralizeAge(%v, %v) unexpected error: %v", tt.birthDate, tt.refDate, err)
			}
			if got.Label != tt.wantLabel {
				t.Errorf("GeneralizeAge(%v, %v).Label = %q, want %q", tt.birthDate, tt.refDate, got.Label, tt.wantLabel)
			}
		})
	}
}

func TestGeneralizeAge_AgeBandFields(t *testing.T) {
	// Verify that MinAge and MaxAge fields are populated correctly for a known band
	now := time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC)
	birthDate := time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC) // age ~10

	got, err := GeneralizeAge(birthDate, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Label != "10-14" {
		t.Errorf("Label = %q, want %q", got.Label, "10-14")
	}
	if got.MinAge != 10 {
		t.Errorf("MinAge = %d, want 10", got.MinAge)
	}
	if got.MaxAge != 14 {
		t.Errorf("MaxAge = %d, want 14", got.MaxAge)
	}
}

func TestGeneralizeAge_80PlusBandMaxAge(t *testing.T) {
	// The 80+ band should have a high conventional ceiling for MaxAge
	now := time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC)
	birthDate := time.Date(1940, 1, 1, 0, 0, 0, 0, time.UTC) // age ~86

	got, err := GeneralizeAge(birthDate, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Label != "80+" {
		t.Errorf("Label = %q, want %q", got.Label, "80+")
	}
	if got.MinAge != 80 {
		t.Errorf("MinAge = %d, want 80", got.MinAge)
	}
	// MaxAge should be set to a conventional ceiling (e.g. 199) per domain_types.go
	if got.MaxAge < 80 {
		t.Errorf("MaxAge = %d, want >= 80 for 80+ band", got.MaxAge)
	}
}

func TestGeneralizeIncome(t *testing.T) {
	// The contract defines income in cents relative to minimum wage.
	// We test with known bands. The exact SM value is a domain constant,
	// so we test boundary behavior using the defined IncomeBand constants.
	tests := []struct {
		name             string
		totalIncomeCents int64
		want             IncomeBand
	}{
		{
			name:             "zero income maps to lowest band",
			totalIncomeCents: 0,
			want:             IncomeBand0to05SM,
		},
		{
			name:             "negative income clamps to lowest band",
			totalIncomeCents: -50000,
			want:             IncomeBand0to05SM,
		},
		{
			name:             "very high income maps to 5+SM",
			totalIncomeCents: 100000000, // 1 million reais in cents
			want:             IncomeBand5PlusSM,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GeneralizeIncome(tt.totalIncomeCents)
			if got != tt.want {
				t.Errorf("GeneralizeIncome(%d) = %q, want %q", tt.totalIncomeCents, got, tt.want)
			}
		})
	}
}

func TestGeneralizeIncome_AllBandsReachable(t *testing.T) {
	// Verify that each defined IncomeBand constant can be produced.
	// We use a range of income values that should cover all bands.
	// The exact thresholds depend on the minimum wage constant in the domain,
	// but we can verify at least the extreme bands are reachable.
	allBands := []IncomeBand{
		IncomeBand0to05SM,
		IncomeBand05to1SM,
		IncomeBand1to2SM,
		IncomeBand2to3SM,
		IncomeBand3to5SM,
		IncomeBand5PlusSM,
	}

	// Generate a wide range of incomes and collect which bands appear
	seen := make(map[IncomeBand]bool)
	for cents := int64(0); cents <= 10_000_00 * 100; cents += 10000 {
		band := GeneralizeIncome(cents)
		seen[band] = true
	}

	for _, b := range allBands {
		if !seen[b] {
			t.Errorf("IncomeBand %q was never produced across income range [0, 10_000_00*100]", b)
		}
	}
}
