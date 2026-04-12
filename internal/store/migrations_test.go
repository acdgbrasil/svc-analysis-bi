package store

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Migration struct validation tests
// ---------------------------------------------------------------------------

func TestValidateMigrations_NoDuplicateVersions(t *testing.T) {
	migrations := []Migration{
		{Version: 1, Name: "create tables", SQL: "CREATE TABLE t1 (id INT);"},
		{Version: 2, Name: "add index", SQL: "CREATE INDEX idx ON t1(id);"},
		{Version: 3, Name: "add column", SQL: "ALTER TABLE t1 ADD COLUMN name TEXT;"},
	}

	err := validateMigrations(migrations)
	if err != nil {
		t.Errorf("validateMigrations() unexpected error: %v", err)
	}
}

func TestValidateMigrations_DuplicateVersionDetected(t *testing.T) {
	migrations := []Migration{
		{Version: 1, Name: "first", SQL: "CREATE TABLE t1 (id INT);"},
		{Version: 2, Name: "second", SQL: "CREATE TABLE t2 (id INT);"},
		{Version: 2, Name: "duplicate of second", SQL: "CREATE TABLE t3 (id INT);"},
	}

	err := validateMigrations(migrations)
	if err == nil {
		t.Fatal("validateMigrations() expected ErrMigrationDuplicate, got nil")
	}
	if err != ErrMigrationDuplicate {
		t.Errorf("validateMigrations() error = %v, want %v", err, ErrMigrationDuplicate)
	}
}

func TestValidateMigrations_EmptyList(t *testing.T) {
	err := validateMigrations(nil)
	if err != nil {
		t.Errorf("validateMigrations(nil) unexpected error: %v", err)
	}

	err = validateMigrations([]Migration{})
	if err != nil {
		t.Errorf("validateMigrations([]) unexpected error: %v", err)
	}
}

func TestValidateMigrations_SingleMigration(t *testing.T) {
	migrations := []Migration{
		{Version: 1, Name: "initial", SQL: "CREATE TABLE t (id INT);"},
	}

	err := validateMigrations(migrations)
	if err != nil {
		t.Errorf("validateMigrations() unexpected error: %v", err)
	}
}

func TestValidateMigrations_ManyDuplicates(t *testing.T) {
	migrations := []Migration{
		{Version: 1, Name: "first", SQL: "SELECT 1;"},
		{Version: 1, Name: "dup-1", SQL: "SELECT 2;"},
		{Version: 1, Name: "dup-2", SQL: "SELECT 3;"},
	}

	err := validateMigrations(migrations)
	if err == nil {
		t.Fatal("validateMigrations() expected error for multiple duplicates, got nil")
	}
}

// ---------------------------------------------------------------------------
// Migration ordering tests
// ---------------------------------------------------------------------------

func TestValidateMigrations_VersionsNotAscending(t *testing.T) {
	tests := []struct {
		name       string
		migrations []Migration
	}{
		{
			name: "descending order",
			migrations: []Migration{
				{Version: 3, Name: "third", SQL: "SELECT 3;"},
				{Version: 2, Name: "second", SQL: "SELECT 2;"},
				{Version: 1, Name: "first", SQL: "SELECT 1;"},
			},
		},
		{
			name: "gap then lower",
			migrations: []Migration{
				{Version: 1, Name: "first", SQL: "SELECT 1;"},
				{Version: 5, Name: "fifth", SQL: "SELECT 5;"},
				{Version: 3, Name: "third", SQL: "SELECT 3;"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// If the implementation validates ordering, this should error.
			// If ordering is enforced at RunMigrations level only, this test
			// documents the expected contract.
			err := validateMigrations(tt.migrations)
			if err == nil {
				t.Error("validateMigrations() expected error for non-ascending versions, got nil")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AllowedGenericDimensions sanity check
// ---------------------------------------------------------------------------

func TestAllowedGenericDimensions_ContainsExpectedTables(t *testing.T) {
	expected := []string{
		"dim_housing_type",
		"dim_education_level",
		"dim_benefit_type",
		"dim_referral_destination",
		"dim_violation_type",
	}

	allowed := AllowedGenericDimensions()
	if len(allowed) != len(expected) {
		t.Fatalf("AllowedGenericDimensions() has %d entries, want %d", len(allowed), len(expected))
	}

	lookup := make(map[string]bool)
	for _, d := range allowed {
		lookup[d] = true
	}

	for _, e := range expected {
		if !lookup[e] {
			t.Errorf("AllowedGenericDimensions missing %q", e)
		}
	}
}

// ---------------------------------------------------------------------------
// Sentinel error identity tests
// ---------------------------------------------------------------------------

func TestSentinelErrors_AreDistinct(t *testing.T) {
	errs := []error{
		ErrConnectionFailed,
		ErrPingFailed,
		ErrMigrationFailed,
		ErrMigrationDuplicate,
		ErrDimensionInsertFailed,
		ErrDimensionNotFound,
		ErrEventAlreadyProcessed,
		ErrDLQInsertFailed,
		ErrEventCheckFailed,
		ErrEventMarkFailed,
	}

	seen := make(map[string]int)
	for i, e := range errs {
		msg := e.Error()
		if prev, ok := seen[msg]; ok {
			t.Errorf("sentinel errors [%d] and [%d] have identical message: %q", prev, i, msg)
		}
		seen[msg] = i
	}
}
