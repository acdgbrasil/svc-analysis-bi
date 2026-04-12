package ingestion

import "encoding/json"

// ---------------------------------------------------------------------------
// Internal JSON deserialization structs matching NATS payload format.
// These are unexported and used only for json.Unmarshal inside the anonymizer.
// ---------------------------------------------------------------------------

type jsonMetadata struct {
	EventID       string `json:"eventId"`
	OccurredAt    string `json:"occurredAt"`
	SchemaVersion string `json:"schemaVersion"`
}

type jsonPatientCreated struct {
	Metadata        jsonMetadata `json:"metadata"`
	PatientID       string       `json:"patientId"`
	PersonID        string       `json:"personId"`
	BirthDate       string       `json:"birthDate"`
	Sex             string       `json:"sex"`
	CEP             string       `json:"cep"`
	HousingType     string       `json:"housingType"`
	TotalIncomeCents *int64      `json:"totalIncomeCents"`
}

type jsonAssessmentUpdated struct {
	Metadata  jsonMetadata     `json:"metadata"`
	PatientID string           `json:"patientId"`
	Before    json.RawMessage  `json:"before"`
	After     json.RawMessage  `json:"after"`
}

type jsonHealthStatusAfter struct {
	ICDCode  string `json:"icdCode"`
	ICDLabel string `json:"icdLabel"`
	Chapter  string `json:"chapter"`
	Block    string `json:"block"`
}

type jsonAppointmentRegistered struct {
	Metadata               jsonMetadata `json:"metadata"`
	PatientID              string       `json:"patientId"`
	AppointmentID          string       `json:"appointmentId"`
	ProfessionalInChargeID string       `json:"professionalInChargeId"`
	AppointmentType        string       `json:"appointmentType"`
}

type jsonReferralCreated struct {
	Metadata           jsonMetadata `json:"metadata"`
	PatientID          string       `json:"patientId"`
	ReferralID         string       `json:"referralId"`
	ReferredPersonID   string       `json:"referredPersonId"`
	DestinationService string       `json:"destinationService"`
	Status             string       `json:"status"`
}

type jsonRightsViolationReported struct {
	Metadata      jsonMetadata `json:"metadata"`
	PatientID     string       `json:"patientId"`
	ReportID      string       `json:"reportId"`
	VictimID      string       `json:"victimId"`
	ViolationType string       `json:"violationType"`
}

type jsonFamilyMemberAdded struct {
	Metadata     jsonMetadata `json:"metadata"`
	PatientID    string       `json:"patientId"`
	MemberID     string       `json:"memberId"`
	Relationship string       `json:"relationship"`
}

type jsonFamilyMemberRemoved struct {
	Metadata  jsonMetadata `json:"metadata"`
	PatientID string       `json:"patientId"`
	MemberID  string       `json:"memberId"`
}

type jsonPrimaryCaregiverAssigned struct {
	Metadata    jsonMetadata `json:"metadata"`
	PatientID   string       `json:"patientId"`
	CaregiverID string       `json:"caregiverId"`
}

// jsonGenericAssessment is used for assessment events that produce a
// SnapshotPayload (housing, education, socioeconomic, etc.). We only
// need metadata and patientId; the After payload is event-specific but
// for snapshot purposes we just need to acknowledge the update.
type jsonGenericAssessment struct {
	Metadata  jsonMetadata    `json:"metadata"`
	PatientID string          `json:"patientId"`
	Before    json.RawMessage `json:"before"`
	After     json.RawMessage `json:"after"`
}
