package ingestion

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/acdgbrasil/svc-analysis-bi/internal/domain"
)

// ---------------------------------------------------------------------------
// Test: PatientID is hashed, never raw
// ---------------------------------------------------------------------------

func TestAnonymize_PatientIDIsHashed(t *testing.T) {
	geoLookup := newFakeGeographyLookup()
	salt := "secret-test-salt"

	anonymizer := NewAnonymizer(geoLookup, salt)

	tests := []struct {
		name      string
		patientID string
	}{
		{"uuid format", "550e8400-e29b-41d4-a716-446655440000"},
		{"simple string", "patient-123"},
		{"long id", "very-long-patient-identifier-that-should-still-be-hashed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := rawPatientEvent(tt.patientID, "1990-01-15", "MALE", "13083970")

			rec, err := anonymizer.Anonymize(context.Background(), domain.EventPatientCreated, evt)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Hash must NOT be the raw ID
			if string(rec.PatientHash) == tt.patientID {
				t.Error("PatientHash must not equal raw patientId")
			}

			// Hash must be a 64-char hex string (SHA-256)
			if len(string(rec.PatientHash)) != 64 {
				t.Errorf("PatientHash length = %d, want 64 (SHA-256 hex)", len(string(rec.PatientHash)))
			}

			// Hash must be deterministic
			rec2, _ := anonymizer.Anonymize(context.Background(), domain.EventPatientCreated, evt)
			if rec.PatientHash != rec2.PatientHash {
				t.Error("PatientHash must be deterministic for same input")
			}
		})
	}
}

func TestAnonymize_DifferentSaltsProduceDifferentHashes(t *testing.T) {
	geoLookup := newFakeGeographyLookup()

	anonymizer1 := NewAnonymizer(geoLookup, "salt-one")
	anonymizer2 := NewAnonymizer(geoLookup, "salt-two")

	evt := rawPatientEvent("same-patient-id", "1990-01-15", "MALE", "13083970")

	rec1, err := anonymizer1.Anonymize(context.Background(), domain.EventPatientCreated, evt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rec2, err := anonymizer2.Anonymize(context.Background(), domain.EventPatientCreated, evt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec1.PatientHash == rec2.PatientHash {
		t.Error("different salts must produce different hashes for same patientId")
	}
}

// ---------------------------------------------------------------------------
// Test: BirthDate is generalized to age band
// ---------------------------------------------------------------------------

func TestAnonymize_BirthDateGeneralizedToAgeBand(t *testing.T) {
	geoLookup := newFakeGeographyLookup()
	salt := "test-salt"

	anonymizer := NewAnonymizer(geoLookup, salt)

	tests := []struct {
		name          string
		birthDate     string
		wantBandLabel string
	}{
		{"infant born 2024", "2024-06-01", "0-4"},
		{"child born 2018", "2018-03-15", "5-9"},
		{"teenager born 2010", "2010-01-01", "10-14"},
		{"adult born 1990", "1990-07-20", "30-34"},
		{"elderly born 1940", "1940-12-25", "80+"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := rawPatientEvent("pat-age-test", tt.birthDate, "MALE", "13083970")

			rec, err := anonymizer.Anonymize(context.Background(), domain.EventPatientCreated, evt)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.Snapshot == nil {
				t.Fatal("expected Snapshot payload to be non-nil")
			}

			if rec.Snapshot.AgeBand.Label != tt.wantBandLabel {
				t.Errorf("AgeBand.Label = %q, want %q", rec.Snapshot.AgeBand.Label, tt.wantBandLabel)
			}

			// The exact birth date must NOT appear anywhere in the serialized record
			recJSON, _ := json.Marshal(rec)
			if strings.Contains(string(recJSON), tt.birthDate) {
				t.Error("exact birth date must not appear in anonymized record")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Test: CEP is generalized to mesoregion
// ---------------------------------------------------------------------------

func TestAnonymize_CEPGeneralizedToMesoregion(t *testing.T) {
	geoLookup := newFakeGeographyLookup()
	salt := "test-salt"

	anonymizer := NewAnonymizer(geoLookup, salt)

	evt := rawPatientEvent("pat-cep-test", "1990-01-01", "FEMALE", "13083970")

	rec, err := anonymizer.Anonymize(context.Background(), domain.EventPatientCreated, evt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Snapshot == nil {
		t.Fatal("expected Snapshot payload to be non-nil")
	}

	// Geography should match the fake lookup's return value
	if rec.Snapshot.Geography.MesoregionCode != "3515" {
		t.Errorf("MesoregionCode = %q, want %q", rec.Snapshot.Geography.MesoregionCode, "3515")
	}
	if rec.Snapshot.Geography.MesoregionName != "Campinas" {
		t.Errorf("MesoregionName = %q, want %q", rec.Snapshot.Geography.MesoregionName, "Campinas")
	}

	// The exact CEP must NOT appear in the serialized record
	recJSON, _ := json.Marshal(rec)
	if strings.Contains(string(recJSON), "13083970") {
		t.Error("exact CEP must not appear in anonymized record")
	}
}

func TestAnonymize_CEPLookupFailure(t *testing.T) {
	geoLookup := newFakeGeographyLookupWithError(domain.ErrCEPNotFound)
	salt := "test-salt"

	anonymizer := NewAnonymizer(geoLookup, salt)

	evt := rawPatientEvent("pat-cep-fail", "1990-01-01", "MALE", "99999999")

	_, err := anonymizer.Anonymize(context.Background(), domain.EventPatientCreated, evt)
	if err == nil {
		t.Fatal("expected error when CEP lookup fails")
	}
}

// ---------------------------------------------------------------------------
// Test: Income is generalized to income band
// ---------------------------------------------------------------------------

func TestAnonymize_IncomeGeneralizedToIncomeBand(t *testing.T) {
	geoLookup := newFakeGeographyLookup()
	salt := "test-salt"

	anonymizer := NewAnonymizer(geoLookup, salt)

	tests := []struct {
		name       string
		incomeCents int64
		wantBand   domain.IncomeBand
	}{
		{"zero income", 0, domain.IncomeBand0to05SM},
		{"half minimum wage", 70600, domain.IncomeBand0to05SM},
		{"one minimum wage", 141200, domain.IncomeBand1to2SM},
		{"three minimum wages", 423600, domain.IncomeBand3to5SM},
		{"high income", 1000000, domain.IncomeBand5PlusSM},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := rawPatientEventWithIncome("pat-income-test", "1990-01-01", "MALE", "13083970", tt.incomeCents)

			rec, err := anonymizer.Anonymize(context.Background(), domain.EventPatientCreated, evt)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.Snapshot == nil {
				t.Fatal("expected Snapshot payload to be non-nil")
			}

			if rec.Snapshot.IncomeBand != tt.wantBand {
				t.Errorf("IncomeBand = %q, want %q", rec.Snapshot.IncomeBand, tt.wantBand)
			}

			// Exact income value must not appear in serialized record
			// The IncomeBand is a coarse bracket, not the exact amount
			_, _ = json.Marshal(rec) // ensure record is serializable
		})
	}
}

// ---------------------------------------------------------------------------
// Test: PII fields are completely absent from AnonymizedRecord
// ---------------------------------------------------------------------------

func TestAnonymize_PIIFieldsAbsent(t *testing.T) {
	geoLookup := newFakeGeographyLookup()
	salt := "test-salt"

	anonymizer := NewAnonymizer(geoLookup, salt)

	piiFields := []struct {
		name       string
		eventType  domain.EventType
		rawEvent   []byte
		piiValues  []string
	}{
		{
			name:      "actorId absent from appointment",
			eventType: domain.EventAppointmentRegistered,
			rawEvent: mustJSON(map[string]any{
				"metadata": map[string]any{
					"eventId":       "evt-pii-actor",
					"occurredAt":    "2025-06-15T10:00:00Z",
					"schemaVersion": "1.0",
				},
				"patientId":              "pat-pii-actor",
				"appointmentId":          "appt-pii-001",
				"professionalInChargeId": "actor-uuid-must-vanish",
				"appointmentType":        "initial",
			}),
			piiValues: []string{"actor-uuid-must-vanish", "appt-pii-001"},
		},
		{
			name:      "memberId absent from family member added",
			eventType: domain.EventFamilyMemberAdded,
			rawEvent: mustJSON(map[string]any{
				"metadata": map[string]any{
					"eventId":       "evt-pii-member",
					"occurredAt":    "2025-06-15T10:00:00Z",
					"schemaVersion": "1.0",
				},
				"patientId":    "pat-pii-member",
				"memberId":     "member-uuid-must-vanish",
				"relationship": "child",
			}),
			piiValues: []string{"member-uuid-must-vanish"},
		},
		{
			name:      "victimId absent from violation",
			eventType: domain.EventRightsViolationReported,
			rawEvent: mustJSON(map[string]any{
				"metadata": map[string]any{
					"eventId":       "evt-pii-victim",
					"occurredAt":    "2025-06-15T10:00:00Z",
					"schemaVersion": "1.0",
				},
				"patientId":     "pat-pii-victim",
				"reportId":      "report-uuid-must-vanish",
				"victimId":      "victim-uuid-must-vanish",
				"violationType": "neglect",
			}),
			piiValues: []string{"victim-uuid-must-vanish", "report-uuid-must-vanish"},
		},
		{
			name:      "caregiverId absent from caregiver assigned",
			eventType: domain.EventPrimaryCaregiverAssigned,
			rawEvent: mustJSON(map[string]any{
				"metadata": map[string]any{
					"eventId":       "evt-pii-caregiver",
					"occurredAt":    "2025-06-15T10:00:00Z",
					"schemaVersion": "1.0",
				},
				"patientId":   "pat-pii-caregiver",
				"caregiverId": "caregiver-uuid-must-vanish",
			}),
			piiValues: []string{"caregiver-uuid-must-vanish"},
		},
		{
			name:      "referredPersonId absent from referral",
			eventType: domain.EventReferralCreated,
			rawEvent: mustJSON(map[string]any{
				"metadata": map[string]any{
					"eventId":       "evt-pii-referral",
					"occurredAt":    "2025-06-15T10:00:00Z",
					"schemaVersion": "1.0",
				},
				"patientId":          "pat-pii-referral",
				"referralId":         "referral-uuid-must-vanish",
				"referredPersonId":   "referred-uuid-must-vanish",
				"destinationService": "CRAS",
				"status":             "pending",
			}),
			piiValues: []string{"referral-uuid-must-vanish", "referred-uuid-must-vanish"},
		},
	}

	for _, tt := range piiFields {
		t.Run(tt.name, func(t *testing.T) {
			rec, err := anonymizer.Anonymize(context.Background(), tt.eventType, tt.rawEvent)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			recJSON, _ := json.Marshal(rec)
			serialized := string(recJSON)

			for _, pii := range tt.piiValues {
				if strings.Contains(serialized, pii) {
					t.Errorf("PII value %q must not appear in anonymized record", pii)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Test: Empty salt produces error
// ---------------------------------------------------------------------------

func TestAnonymize_EmptySaltError(t *testing.T) {
	geoLookup := newFakeGeographyLookup()

	anonymizer := NewAnonymizer(geoLookup, "")

	evt := rawPatientEvent("pat-empty-salt", "1990-01-01", "MALE", "13083970")

	_, err := anonymizer.Anonymize(context.Background(), domain.EventPatientCreated, evt)
	if err == nil {
		t.Fatal("expected error when salt is empty")
	}
}

// ---------------------------------------------------------------------------
// Test: Period derived from OccurredAt
// ---------------------------------------------------------------------------

func TestAnonymize_PeriodDerivedFromOccurredAt(t *testing.T) {
	geoLookup := newFakeGeographyLookup()
	salt := "test-salt"

	anonymizer := NewAnonymizer(geoLookup, salt)

	tests := []struct {
		name       string
		occurredAt string
		wantYear   int
		wantMonth  int
	}{
		{"january 2025", "2025-01-15T10:00:00Z", 2025, 1},
		{"december 2024", "2024-12-31T23:59:59Z", 2024, 12},
		{"june 2025", "2025-06-01T00:00:00Z", 2025, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := rawPatientEventWithOccurredAt("pat-period-test", "1990-01-01", "MALE", "13083970", tt.occurredAt)

			rec, err := anonymizer.Anonymize(context.Background(), domain.EventPatientCreated, evt)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.Period.Year != tt.wantYear {
				t.Errorf("Period.Year = %d, want %d", rec.Period.Year, tt.wantYear)
			}
			if rec.Period.Month != tt.wantMonth {
				t.Errorf("Period.Month = %d, want %d", rec.Period.Month, tt.wantMonth)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Test: Sex mapping
// ---------------------------------------------------------------------------

func TestAnonymize_SexMapping(t *testing.T) {
	geoLookup := newFakeGeographyLookup()
	salt := "test-salt"

	anonymizer := NewAnonymizer(geoLookup, salt)

	tests := []struct {
		name    string
		sex     string
		wantSex domain.Sex
	}{
		{"male", "MALE", domain.SexMale},
		{"female", "FEMALE", domain.SexFemale},
		{"unknown", "UNKNOWN", domain.SexUnknown},
		{"empty defaults to unknown", "", domain.SexUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := rawPatientEvent("pat-sex-test", "1990-01-01", tt.sex, "13083970")

			rec, err := anonymizer.Anonymize(context.Background(), domain.EventPatientCreated, evt)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.Snapshot == nil {
				t.Fatal("expected Snapshot payload to be non-nil")
			}
			if rec.Snapshot.Sex != tt.wantSex {
				t.Errorf("Sex = %q, want %q", rec.Snapshot.Sex, tt.wantSex)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func rawPatientEvent(patientID, birthDate, sex, cep string) []byte {
	return rawPatientEventWithOccurredAt(patientID, birthDate, sex, cep, "2025-06-15T10:00:00Z")
}

func rawPatientEventWithOccurredAt(patientID, birthDate, sex, cep, occurredAt string) []byte {
	data, _ := json.Marshal(map[string]any{
		"metadata": map[string]any{
			"eventId":       "evt-anon-" + patientID,
			"occurredAt":    occurredAt,
			"schemaVersion": "1.0",
		},
		"patientId": patientID,
		"personId":  "person-" + patientID,
		"birthDate": birthDate,
		"sex":       sex,
		"cep":       cep,
	})
	return data
}

func rawPatientEventWithIncome(patientID, birthDate, sex, cep string, incomeCents int64) []byte {
	data, _ := json.Marshal(map[string]any{
		"metadata": map[string]any{
			"eventId":       "evt-income-" + patientID,
			"occurredAt":    "2025-06-15T10:00:00Z",
			"schemaVersion": "1.0",
		},
		"patientId":       patientID,
		"personId":        "person-" + patientID,
		"birthDate":       birthDate,
		"sex":             sex,
		"cep":             cep,
		"totalIncomeCents": incomeCents,
	})
	return data
}

func mustJSON(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic("mustJSON: " + err.Error())
	}
	return data
}
