package domain

// EventType enumerates all NATS subjects this service subscribes to.
type EventType string

const (
	EventPatientCreated           EventType = "social-care.patient.created"
	EventFamilyMemberAdded        EventType = "social-care.family-member.added"
	EventFamilyMemberRemoved      EventType = "social-care.family-member.removed"
	EventPrimaryCaregiverAssigned EventType = "social-care.primary-caregiver.assigned"
	EventSocialIdentityUpdated    EventType = "social-care.social-identity.updated"
	EventHousingConditionUpdated  EventType = "social-care.housing-condition.updated"
	EventSocioEconomicUpdated     EventType = "social-care.socioeconomic-situation.updated"
	EventWorkAndIncomeUpdated     EventType = "social-care.work-and-income.updated"
	EventEducationalStatusUpdated EventType = "social-care.educational-status.updated"
	EventHealthStatusUpdated      EventType = "social-care.health-status.updated"
	EventCommunitySupportUpdated  EventType = "social-care.community-support-network.updated"
	EventSocialHealthSummaryUpdated EventType = "social-care.social-health-summary.updated"
	EventAppointmentRegistered    EventType = "social-care.appointment.registered"
	EventIntakeInfoUpdated        EventType = "social-care.intake-info.updated"
	EventPlacementHistoryUpdated  EventType = "social-care.placement-history.updated"
	EventRightsViolationReported  EventType = "social-care.rights-violation.reported"
	EventReferralCreated          EventType = "social-care.referral.created"
)

// EventMetadata contains traceability fields present in every event.
type EventMetadata struct {
	EventID       string
	OccurredAt    string
	SchemaVersion string
}

// PatientCreatedEvent is emitted when a new patient record is created.
type PatientCreatedEvent struct {
	Metadata  EventMetadata
	PatientID string
	PersonID  string
}

// FamilyMemberAddedEvent is emitted when a family member is added.
type FamilyMemberAddedEvent struct {
	Metadata     EventMetadata
	PatientID    string
	MemberID     string
	Relationship string
}

// FamilyMemberRemovedEvent is emitted when a family member is removed.
type FamilyMemberRemovedEvent struct {
	Metadata  EventMetadata
	PatientID string
	MemberID  string
}

// PrimaryCaregiverAssignedEvent is emitted when a primary caregiver is assigned.
type PrimaryCaregiverAssignedEvent struct {
	Metadata    EventMetadata
	PatientID   string
	CaregiverID string
}

// AssessmentUpdatedEvent is the generic payload for all assessment updates.
// Before/After are kept as raw JSON bytes because the concrete assessment
// type varies by NATS subject (10 different schemas). The ingestion layer
// deserializes them into typed structs based on the EventType. Using
// json.RawMessage avoids map[string]any while preserving deferred decoding.
type AssessmentUpdatedEvent struct {
	Metadata  EventMetadata
	PatientID string
	Before    []byte // json.RawMessage — nil on first assignment
	After     []byte // json.RawMessage — full assessment snapshot
}

// AppointmentRegisteredEvent is emitted when a social care appointment is registered.
type AppointmentRegisteredEvent struct {
	Metadata               EventMetadata
	PatientID              string
	AppointmentID          string
	ProfessionalInChargeID string
	AppointmentType        string
}

// RightsViolationReportedEvent is emitted when a rights violation report is filed.
type RightsViolationReportedEvent struct {
	Metadata      EventMetadata
	PatientID     string
	ReportID      string
	VictimID      string
	ViolationType string
}

// ReferralCreatedEvent is emitted when a referral is created.
type ReferralCreatedEvent struct {
	Metadata           EventMetadata
	PatientID          string
	ReferralID         string
	ReferredPersonID   string
	DestinationService string
	Status             string
}
