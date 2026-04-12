package ingestion

import "encoding/json"

// ---------------------------------------------------------------------------
// Internal JSON deserialization structs matching the NATS payload format
// published by svc-social-care's NATSEventPublisher (Swift/Codable).
//
// Swift serializes events as FLAT objects with top-level fields:
//   - "id" (UUID string) — event identifier
//   - "occurredAt" (ISO 8601) — timestamp
//   - "patientId", "actorId", etc. — camelCase field names
//
// There is NO "metadata" wrapper — id and occurredAt are top-level.
// The "actorId" field is present in all events but discarded (PII).
// ---------------------------------------------------------------------------

// jsonEventBase contains the fields common to ALL events from svc-social-care.
// Swift's DomainEvent protocol requires id, occurredAt, and actorId.
type jsonEventBase struct {
	ID         string `json:"id"`
	OccurredAt string `json:"occurredAt"`
	ActorID    string `json:"actorId"` // PII — always discarded
	PatientID  string `json:"patientId"`
}

type jsonPatientCreated struct {
	jsonEventBase
	PersonID         string `json:"personId"`
	BirthDate        string `json:"birthDate"`        // optional — may come in SocialIdentityUpdated instead
	Sex              string `json:"sex"`               // optional
	CEP              string `json:"cep"`               // optional
	HousingType      string `json:"housingType"`       // optional
	TotalIncomeCents *int64 `json:"totalIncomeCents"`  // optional
}

type jsonAssessmentUpdated struct {
	jsonEventBase
	Before json.RawMessage `json:"before"`
	After  json.RawMessage `json:"after"`
}

type jsonHealthStatusAfter struct {
	ICDCode  string `json:"icdCode"`
	ICDLabel string `json:"icdLabel"`
	Chapter  string `json:"chapter"`
	Block    string `json:"block"`
}

type jsonAppointmentRegistered struct {
	jsonEventBase
	AppointmentID          string `json:"appointmentId"`
	ProfessionalInChargeID string `json:"professionalInChargeId"`
	Type                   string `json:"type"` // Swift uses "type", not "appointmentType"
}

type jsonReferralCreated struct {
	jsonEventBase
	ReferralID         string `json:"referralId"`
	ReferredPersonID   string `json:"referredPersonId"`
	DestinationService string `json:"destinationService"`
	Status             string `json:"status"`
}

type jsonRightsViolationReported struct {
	jsonEventBase
	ReportID      string `json:"reportId"`
	VictimID      string `json:"victimId"`
	ViolationType string `json:"violationType"`
}

type jsonFamilyMemberAdded struct {
	jsonEventBase
	MemberID     string `json:"memberId"`
	Relationship string `json:"relationship"`
}

type jsonFamilyMemberRemoved struct {
	jsonEventBase
	MemberID string `json:"memberId"`
}

type jsonPrimaryCaregiverAssigned struct {
	jsonEventBase
	CaregiverID string `json:"caregiverId"`
}

// jsonGenericAssessment is used for assessment events that produce a
// SnapshotPayload (housing, education, socioeconomic, etc.).
type jsonGenericAssessment struct {
	jsonEventBase
	Before json.RawMessage `json:"before"`
	After  json.RawMessage `json:"after"`
}
