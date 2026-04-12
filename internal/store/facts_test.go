package store

import (
	"errors"
	"testing"

	"github.com/acdgbrasil/svc-analysis-bi/internal/domain"
	"github.com/acdgbrasil/svc-analysis-bi/internal/ingestion"
)

func TestPgFactStoreImplementsInterface(t *testing.T) {
	// Compile-time check is in facts.go via var _ ingestion.FactStore = ...
	// This test verifies the type assertion at runtime.
	var fs ingestion.FactStore = &PgFactStore{}
	if fs == nil {
		t.Fatal("PgFactStore does not satisfy ingestion.FactStore")
	}
}

func TestNewFactStoreCreation(t *testing.T) {
	// NewFactStore must not panic with a nil pool. The pool is only used
	// when methods are actually called, not at construction time.
	fs := NewFactStore(nil)
	if fs == nil {
		t.Fatal("expected non-nil PgFactStore")
	}
	if fs.dims == nil {
		t.Fatal("expected non-nil internal PgDimensionStore")
	}
}

func TestNilPayloadErrors(t *testing.T) {
	fs := &PgFactStore{}

	tests := []struct {
		name   string
		kind   ingestion.FactKind
		fn     func() error
	}{
		{
			name: "UpsertPatientSnapshot nil Snapshot",
			kind: ingestion.FactKindPatientSnapshot,
			fn: func() error {
				return fs.UpsertPatientSnapshot(t.Context(), ingestion.AnonymizedRecord{
					Kind: ingestion.FactKindPatientSnapshot,
				})
			},
		},
		{
			name: "IncrementDiagnosis nil Diagnosis",
			kind: ingestion.FactKindDiagnosis,
			fn: func() error {
				return fs.IncrementDiagnosis(t.Context(), ingestion.AnonymizedRecord{
					Kind: ingestion.FactKindDiagnosis,
				})
			},
		},
		{
			name: "IncrementAppointment nil Appointment",
			kind: ingestion.FactKindAppointment,
			fn: func() error {
				return fs.IncrementAppointment(t.Context(), ingestion.AnonymizedRecord{
					Kind: ingestion.FactKindAppointment,
				})
			},
		},
		{
			name: "IncrementReferral nil Referral",
			kind: ingestion.FactKindReferral,
			fn: func() error {
				return fs.IncrementReferral(t.Context(), ingestion.AnonymizedRecord{
					Kind: ingestion.FactKindReferral,
				})
			},
		},
		{
			name: "IncrementViolation nil Violation",
			kind: ingestion.FactKindViolation,
			fn: func() error {
				return fs.IncrementViolation(t.Context(), ingestion.AnonymizedRecord{
					Kind: ingestion.FactKindViolation,
				})
			},
		},
		{
			name: "IncrementBenefit nil Benefit",
			kind: ingestion.FactKindBenefit,
			fn: func() error {
				return fs.IncrementBenefit(t.Context(), ingestion.AnonymizedRecord{
					Kind: ingestion.FactKindBenefit,
				})
			},
		},
		{
			name: "UpsertFamilyComposition nil FamilyComposition",
			kind: ingestion.FactKindFamilyComposition,
			fn: func() error {
				return fs.UpsertFamilyComposition(t.Context(), ingestion.AnonymizedRecord{
					Kind: ingestion.FactKindFamilyComposition,
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err == nil {
				t.Fatal("expected error for nil payload, got nil")
			}
			if !errors.Is(err, ErrNilPayload) {
				t.Fatalf("expected ErrNilPayload, got: %v", err)
			}
		})
	}
}

func TestFactStoreSentinelErrors(t *testing.T) {
	// Verify sentinel errors are distinct and have correct prefixes.
	sentinels := []error{
		ErrFactInsertFailed,
		ErrFactUpsertFailed,
		ErrNilPayload,
		ErrMissingDimensions,
	}

	seen := make(map[string]bool, len(sentinels))
	for _, err := range sentinels {
		msg := err.Error()
		if seen[msg] {
			t.Fatalf("duplicate sentinel error message: %s", msg)
		}
		seen[msg] = true

		if msg == "" {
			t.Fatal("sentinel error has empty message")
		}
	}
}

func TestFactKindPayloadMismatch(t *testing.T) {
	// When record.Kind says one thing but the corresponding payload field is
	// populated on a different field, the method should return ErrNilPayload.
	fs := &PgFactStore{}

	// Snapshot method with a Diagnosis payload only.
	record := ingestion.AnonymizedRecord{
		Kind:   ingestion.FactKindPatientSnapshot,
		Period: domain.Period{Year: 2025, Month: 3},
		Diagnosis: &ingestion.DiagnosisPayload{
			ICDCode: "Q90.0",
		},
	}
	err := fs.UpsertPatientSnapshot(t.Context(), record)
	if !errors.Is(err, ErrNilPayload) {
		t.Fatalf("expected ErrNilPayload for mismatched kind/payload, got: %v", err)
	}

	// Diagnosis method with a Snapshot payload only.
	record2 := ingestion.AnonymizedRecord{
		Kind:   ingestion.FactKindDiagnosis,
		Period: domain.Period{Year: 2025, Month: 3},
		Snapshot: &ingestion.SnapshotPayload{
			AgeBand: domain.AgeBand{Label: "0-4"},
		},
	}
	err = fs.IncrementDiagnosis(t.Context(), record2)
	if !errors.Is(err, ErrNilPayload) {
		t.Fatalf("expected ErrNilPayload for mismatched kind/payload, got: %v", err)
	}
}

func TestFactStoreErrorMessages(t *testing.T) {
	// Verify that all sentinel errors have the "store:" prefix for consistency.
	sentinels := []error{
		ErrFactInsertFailed,
		ErrFactUpsertFailed,
		ErrNilPayload,
		ErrMissingDimensions,
	}

	for _, err := range sentinels {
		msg := err.Error()
		if len(msg) < 7 || msg[:6] != "store:" {
			t.Fatalf("sentinel error %q does not start with 'store:' prefix", msg)
		}
	}
}
