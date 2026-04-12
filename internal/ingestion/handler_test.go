package ingestion

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/acdgbrasil/svc-analysis-bi/internal/domain"
)

// ---------------------------------------------------------------------------
// Test: PatientCreated handler
// ---------------------------------------------------------------------------

func TestHandlePatientCreated_ValidJSON(t *testing.T) {
	geoLookup := newFakeGeographyLookup()
	salt := "test-salt-for-hashing"
	registry := NewEventHandlerRegistry(geoLookup, salt)

	handler, ok := registry[domain.EventPatientCreated]
	if !ok {
		t.Fatal("expected handler registered for EventPatientCreated")
	}

	tests := []struct {
		name       string
		input      map[string]any
		wantKind   FactKind
		wantNoErr  bool
	}{
		{
			name: "valid patient created event with all fields",
			input: map[string]any{
				"id":         "evt-001",
				"occurredAt": "2025-06-15T10:00:00Z",
				"actorId":    "actor-001",
				"patientId":  "pat-uuid-123",
				"personId":   "person-uuid-456",
				"birthDate":  "1990-03-15",
				"sex":        "MALE",
				"cep":        "13083970",
				"housingType": "apartment",
			},
			wantKind:  FactKindPatientSnapshot,
			wantNoErr: true,
		},
		{
			name: "valid patient created event female",
			input: map[string]any{
				"id":         "evt-002",
				"occurredAt": "2025-06-15T10:00:00Z",
				"actorId":    "actor-001",
				"patientId":  "pat-uuid-789",
				"personId":   "person-uuid-012",
				"birthDate":  "2000-01-01",
				"sex":        "FEMALE",
				"cep":        "01310100",
			},
			wantKind:  FactKindPatientSnapshot,
			wantNoErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("failed to marshal test input: %v", err)
			}

			rec, err := handler(context.Background(), data)
			if tt.wantNoErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.wantNoErr && err == nil {
				t.Fatal("expected error but got nil")
			}
			if tt.wantNoErr {
				if rec.Kind != tt.wantKind {
					t.Errorf("Kind = %q, want %q", rec.Kind, tt.wantKind)
				}
				if rec.Snapshot == nil {
					t.Error("expected Snapshot payload to be non-nil")
				}
				if rec.EventID == "" {
					t.Error("expected EventID to be set")
				}
			}
		})
	}
}

func TestHandlePatientCreated_PIIDiscarded(t *testing.T) {
	geoLookup := newFakeGeographyLookup()
	salt := "test-salt-for-hashing"
	registry := NewEventHandlerRegistry(geoLookup, salt)

	handler := registry[domain.EventPatientCreated]

	input := map[string]any{
		"id":         "evt-pii-001",
		"occurredAt": "2025-06-15T10:00:00Z",
		"actorId":    "actor-001",
		"patientId":  "pat-uuid-should-be-hashed",
		"personId":   "person-uuid-should-be-discarded",
		"birthDate":  "1985-07-20",
		"sex":        "MALE",
		"cep":        "13083970",
	}

	data, _ := json.Marshal(input)
	rec, err := handler(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// PatientHash must NOT be the raw patientId
	if string(rec.PatientHash) == "pat-uuid-should-be-hashed" {
		t.Error("PatientHash must be a hash, not the raw patientId")
	}
	if rec.PatientHash == "" {
		t.Error("PatientHash must not be empty")
	}

	// Verify the hash is a hex-encoded SHA-256 (64 chars)
	if len(string(rec.PatientHash)) != 64 {
		t.Errorf("PatientHash length = %d, want 64 (SHA-256 hex)", len(string(rec.PatientHash)))
	}
}

func TestHandlePatientCreated_InvalidJSON(t *testing.T) {
	geoLookup := newFakeGeographyLookup()
	salt := "test-salt"
	registry := NewEventHandlerRegistry(geoLookup, salt)

	handler := registry[domain.EventPatientCreated]

	tests := []struct {
		name    string
		data    []byte
		wantErr error
	}{
		{
			name:    "completely invalid JSON",
			data:    []byte(`{not-json`),
			wantErr: ErrDeserializationFailed,
		},
		{
			name:    "empty bytes",
			data:    []byte{},
			wantErr: ErrDeserializationFailed,
		},
		{
			name:    "null JSON",
			data:    []byte(`null`),
			wantErr: ErrDeserializationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler(context.Background(), tt.data)
			if err == nil {
				t.Fatal("expected error but got nil")
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want wrapping %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandlePatientCreated_MissingRequiredFields(t *testing.T) {
	geoLookup := newFakeGeographyLookup()
	salt := "test-salt"
	registry := NewEventHandlerRegistry(geoLookup, salt)

	handler := registry[domain.EventPatientCreated]

	tests := []struct {
		name  string
		input map[string]any
	}{
		{
			name: "missing patientId",
			input: map[string]any{
				"id":         "evt-001",
				"occurredAt": "2025-06-15T10:00:00Z",
				"actorId":    "actor-001",
				"personId":   "p-001",
				"birthDate":  "1990-01-01",
				"sex":        "MALE",
				"cep":        "13083970",
			},
		},
		{
			name: "missing id",
			input: map[string]any{
				"occurredAt": "2025-06-15T10:00:00Z",
				"actorId":    "actor-001",
				"patientId":  "pat-001",
				"personId":   "person-001",
				"birthDate":  "1990-01-01",
				"sex":        "MALE",
				"cep":        "13083970",
			},
		},
		// birthDate is optional — missing birthDate is NOT an error
		// (demographics may arrive via SocialIdentityUpdatedEvent instead)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := json.Marshal(tt.input)
			_, err := handler(context.Background(), data)
			if err == nil {
				t.Fatal("expected error for missing required field")
			}
			if !errors.Is(err, ErrAnonymizationFailed) && !errors.Is(err, ErrDeserializationFailed) {
				t.Errorf("error = %v, want ErrAnonymizationFailed or ErrDeserializationFailed", err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Test: HealthStatusUpdated handler (AssessmentUpdatedEvent)
// ---------------------------------------------------------------------------

func TestHandleHealthStatusUpdated_ValidJSON(t *testing.T) {
	geoLookup := newFakeGeographyLookup()
	salt := "test-salt"
	registry := NewEventHandlerRegistry(geoLookup, salt)

	handler, ok := registry[domain.EventHealthStatusUpdated]
	if !ok {
		t.Fatal("expected handler registered for EventHealthStatusUpdated")
	}

	afterPayload := map[string]any{
		"icdCode":  "E10",
		"icdLabel": "Type 1 diabetes mellitus",
		"chapter":  "IV",
		"block":    "E10-E14",
	}
	afterBytes, _ := json.Marshal(afterPayload)

	input := map[string]any{
		"id":         "evt-health-001",
		"occurredAt": "2025-07-01T14:00:00Z",
		"actorId":    "actor-001",
		"patientId":  "pat-uuid-health",
		"before":     nil,
		"after":      json.RawMessage(afterBytes),
	}

	data, _ := json.Marshal(input)
	rec, err := handler(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Kind != FactKindDiagnosis {
		t.Errorf("Kind = %q, want %q", rec.Kind, FactKindDiagnosis)
	}
	if rec.Diagnosis == nil {
		t.Error("expected Diagnosis payload to be non-nil")
	}
	if rec.Diagnosis != nil && rec.Diagnosis.ICDCode != "E10" {
		t.Errorf("ICDCode = %q, want %q", rec.Diagnosis.ICDCode, "E10")
	}

	// PII check: patientId must be hashed
	if string(rec.PatientHash) == "pat-uuid-health" {
		t.Error("PatientHash must not be raw patientId")
	}
}

func TestHandleHealthStatusUpdated_InvalidJSON(t *testing.T) {
	geoLookup := newFakeGeographyLookup()
	salt := "test-salt"
	registry := NewEventHandlerRegistry(geoLookup, salt)

	handler := registry[domain.EventHealthStatusUpdated]

	_, err := handler(context.Background(), []byte(`{broken`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !errors.Is(err, ErrDeserializationFailed) {
		t.Errorf("error = %v, want wrapping %v", err, ErrDeserializationFailed)
	}
}

// ---------------------------------------------------------------------------
// Test: AppointmentRegistered handler
// ---------------------------------------------------------------------------

func TestHandleAppointmentRegistered_ValidJSON(t *testing.T) {
	geoLookup := newFakeGeographyLookup()
	salt := "test-salt"
	registry := NewEventHandlerRegistry(geoLookup, salt)

	handler, ok := registry[domain.EventAppointmentRegistered]
	if !ok {
		t.Fatal("expected handler registered for EventAppointmentRegistered")
	}

	input := map[string]any{
		"id":                     "evt-appt-001",
		"occurredAt":             "2025-08-10T09:30:00Z",
		"actorId":                "actor-001",
		"patientId":              "pat-uuid-appt",
		"appointmentId":          "appt-uuid-001",
		"professionalInChargeId": "prof-uuid-should-be-discarded",
		"type":                   "initial_assessment",
	}

	data, _ := json.Marshal(input)
	rec, err := handler(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Kind != FactKindAppointment {
		t.Errorf("Kind = %q, want %q", rec.Kind, FactKindAppointment)
	}
	if rec.Appointment == nil {
		t.Error("expected Appointment payload to be non-nil")
	}
	if rec.Appointment != nil && rec.Appointment.AppointmentType != "initial_assessment" {
		t.Errorf("AppointmentType = %q, want %q", rec.Appointment.AppointmentType, "initial_assessment")
	}
}

func TestHandleAppointmentRegistered_PIIAbsent(t *testing.T) {
	geoLookup := newFakeGeographyLookup()
	salt := "test-salt"
	registry := NewEventHandlerRegistry(geoLookup, salt)

	handler := registry[domain.EventAppointmentRegistered]

	input := map[string]any{
		"id":                     "evt-appt-pii",
		"occurredAt":             "2025-08-10T09:30:00Z",
		"actorId":                "actor-001",
		"patientId":              "pat-uuid-appt-raw",
		"appointmentId":          "appt-uuid-raw",
		"professionalInChargeId": "prof-uuid-raw",
		"type":                   "follow_up",
	}

	data, _ := json.Marshal(input)
	rec, err := handler(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// PatientHash must not contain raw patientId
	if string(rec.PatientHash) == "pat-uuid-appt-raw" {
		t.Error("PatientHash must be a hash, not raw patientId")
	}

	// The record should NOT contain professionalInChargeId anywhere
	// We verify by checking the serialized output does not contain the raw ID
	recJSON, _ := json.Marshal(rec)
	if strings.Contains(string(recJSON), "prof-uuid-raw") {
		t.Error("professionalInChargeId must be discarded from the anonymized record")
	}
	if strings.Contains(string(recJSON), "appt-uuid-raw") {
		t.Error("appointmentId must be discarded from the anonymized record")
	}
}

// ---------------------------------------------------------------------------
// Test: ReferralCreated handler
// ---------------------------------------------------------------------------

func TestHandleReferralCreated_ValidJSON(t *testing.T) {
	geoLookup := newFakeGeographyLookup()
	salt := "test-salt"
	registry := NewEventHandlerRegistry(geoLookup, salt)

	handler, ok := registry[domain.EventReferralCreated]
	if !ok {
		t.Fatal("expected handler registered for EventReferralCreated")
	}

	input := map[string]any{
		"id":                 "evt-ref-001",
		"occurredAt":         "2025-09-01T11:00:00Z",
		"actorId":            "actor-001",
		"patientId":          "pat-uuid-ref",
		"referralId":         "ref-uuid-001",
		"referredPersonId":   "person-uuid-referred",
		"destinationService": "CRAS",
		"status":             "pending",
	}

	data, _ := json.Marshal(input)
	rec, err := handler(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Kind != FactKindReferral {
		t.Errorf("Kind = %q, want %q", rec.Kind, FactKindReferral)
	}
	if rec.Referral == nil {
		t.Error("expected Referral payload to be non-nil")
	}
	if rec.Referral != nil && rec.Referral.DestinationService != "CRAS" {
		t.Errorf("DestinationService = %q, want %q", rec.Referral.DestinationService, "CRAS")
	}

	// PII suppression: referralId and referredPersonId must not appear
	recJSON, _ := json.Marshal(rec)
	if strings.Contains(string(recJSON), "ref-uuid-001") {
		t.Error("referralId must be discarded from anonymized record")
	}
	if strings.Contains(string(recJSON), "person-uuid-referred") {
		t.Error("referredPersonId must be discarded from anonymized record")
	}
}

func TestHandleReferralCreated_InvalidJSON(t *testing.T) {
	geoLookup := newFakeGeographyLookup()
	salt := "test-salt"
	registry := NewEventHandlerRegistry(geoLookup, salt)

	handler := registry[domain.EventReferralCreated]

	_, err := handler(context.Background(), []byte(`not-json-at-all`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !errors.Is(err, ErrDeserializationFailed) {
		t.Errorf("error = %v, want wrapping %v", err, ErrDeserializationFailed)
	}
}

// ---------------------------------------------------------------------------
// Test: RightsViolationReported handler
// ---------------------------------------------------------------------------

func TestHandleRightsViolationReported_ValidJSON(t *testing.T) {
	geoLookup := newFakeGeographyLookup()
	salt := "test-salt"
	registry := NewEventHandlerRegistry(geoLookup, salt)

	handler, ok := registry[domain.EventRightsViolationReported]
	if !ok {
		t.Fatal("expected handler registered for EventRightsViolationReported")
	}

	input := map[string]any{
		"id":             "evt-viol-001",
		"occurredAt":     "2025-10-05T16:00:00Z",
		"actorId":        "actor-001",
		"patientId":      "pat-uuid-viol",
		"reportId":       "report-uuid-001",
		"victimId":       "victim-uuid-should-be-discarded",
		"violationType":  "physical_abuse",
	}

	data, _ := json.Marshal(input)
	rec, err := handler(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Kind != FactKindViolation {
		t.Errorf("Kind = %q, want %q", rec.Kind, FactKindViolation)
	}
	if rec.Violation == nil {
		t.Error("expected Violation payload to be non-nil")
	}
	if rec.Violation != nil && rec.Violation.ViolationType != "physical_abuse" {
		t.Errorf("ViolationType = %q, want %q", rec.Violation.ViolationType, "physical_abuse")
	}
}

func TestHandleRightsViolationReported_PIIAbsent(t *testing.T) {
	geoLookup := newFakeGeographyLookup()
	salt := "test-salt"
	registry := NewEventHandlerRegistry(geoLookup, salt)

	handler := registry[domain.EventRightsViolationReported]

	input := map[string]any{
		"id":             "evt-viol-pii",
		"occurredAt":     "2025-10-05T16:00:00Z",
		"actorId":        "actor-001",
		"patientId":      "pat-uuid-viol-raw",
		"reportId":       "report-uuid-raw",
		"victimId":       "victim-uuid-raw",
		"violationType":  "neglect",
	}

	data, _ := json.Marshal(input)
	rec, err := handler(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// victimId and reportId must be completely absent
	recJSON, _ := json.Marshal(rec)
	if strings.Contains(string(recJSON), "victim-uuid-raw") {
		t.Error("victimId must be discarded from anonymized record")
	}
	if strings.Contains(string(recJSON), "report-uuid-raw") {
		t.Error("reportId must be discarded from anonymized record")
	}
	if string(rec.PatientHash) == "pat-uuid-viol-raw" {
		t.Error("PatientHash must be hashed, not raw patientId")
	}
}

// ---------------------------------------------------------------------------
// Test: FamilyMemberAdded handler
// ---------------------------------------------------------------------------

func TestHandleFamilyMemberAdded_ValidJSON(t *testing.T) {
	geoLookup := newFakeGeographyLookup()
	salt := "test-salt"
	registry := NewEventHandlerRegistry(geoLookup, salt)

	handler, ok := registry[domain.EventFamilyMemberAdded]
	if !ok {
		t.Fatal("expected handler registered for EventFamilyMemberAdded")
	}

	input := map[string]any{
		"id":           "evt-fam-001",
		"occurredAt":   "2025-11-01T08:00:00Z",
		"actorId":      "actor-001",
		"patientId":    "pat-uuid-fam",
		"memberId":     "member-uuid-should-be-discarded",
		"relationship": "spouse",
	}

	data, _ := json.Marshal(input)
	rec, err := handler(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Kind != FactKindFamilyComposition {
		t.Errorf("Kind = %q, want %q", rec.Kind, FactKindFamilyComposition)
	}
	if rec.FamilyComposition == nil {
		t.Error("expected FamilyComposition payload to be non-nil")
	}
	if rec.FamilyComposition != nil {
		if !rec.FamilyComposition.IsAddition {
			t.Error("expected IsAddition = true for family member added")
		}
		if rec.FamilyComposition.MemberRelationship != "spouse" {
			t.Errorf("MemberRelationship = %q, want %q", rec.FamilyComposition.MemberRelationship, "spouse")
		}
	}

	// memberId must be discarded
	recJSON, _ := json.Marshal(rec)
	if strings.Contains(string(recJSON), "member-uuid-should-be-discarded") {
		t.Error("memberId must be discarded from anonymized record")
	}
}

// ---------------------------------------------------------------------------
// Test: Registry completeness -- all 17 event types have handlers
// ---------------------------------------------------------------------------

func TestEventHandlerRegistry_AllEventsRegistered(t *testing.T) {
	geoLookup := newFakeGeographyLookup()
	salt := "test-salt"
	registry := NewEventHandlerRegistry(geoLookup, salt)

	requiredEvents := []domain.EventType{
		domain.EventPatientCreated,
		domain.EventFamilyMemberAdded,
		domain.EventFamilyMemberRemoved,
		domain.EventPrimaryCaregiverAssigned,
		domain.EventSocialIdentityUpdated,
		domain.EventHousingConditionUpdated,
		domain.EventSocioEconomicUpdated,
		domain.EventWorkAndIncomeUpdated,
		domain.EventEducationalStatusUpdated,
		domain.EventHealthStatusUpdated,
		domain.EventCommunitySupportUpdated,
		domain.EventSocialHealthSummaryUpdated,
		domain.EventAppointmentRegistered,
		domain.EventIntakeInfoUpdated,
		domain.EventPlacementHistoryUpdated,
		domain.EventRightsViolationReported,
		domain.EventReferralCreated,
	}

	for _, evt := range requiredEvents {
		t.Run(string(evt), func(t *testing.T) {
			if _, ok := registry[evt]; !ok {
				t.Errorf("missing handler for event type %q", evt)
			}
		})
	}
}
