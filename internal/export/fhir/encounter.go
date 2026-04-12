package fhir

import "fmt"

// NewEncounter creates a FHIR Encounter resource from appointment data.
// encounterType: type of the appointment
// classCode: encounter class (e.g. "AMB" for ambulatory)
// periodStart/periodEnd: encounter period
// patientRef: reference to the Patient resource
func NewEncounter(id, encounterType, classCode, periodStart, periodEnd, patientRef string) Resource {
	return Resource{
		ResourceType: "Encounter",
		ID:           id,
		Meta: &Meta{
			Profile: []string{ProfileEncounter},
		},
		Status: "finished",
		Class: &Coding{
			System:  SystemEncounter,
			Code:    classCode,
			Display: classCode,
		},
		Type: []CodeableConcept{
			{
				Text: encounterType,
			},
		},
		Subject: &Reference{
			Reference: patientRef,
		},
		Period: &Period{
			Start: periodStart,
			End:   periodEnd,
		},
	}
}

// EncounterFullURL returns a URN for an encounter resource.
func EncounterFullURL(id string) string {
	return fmt.Sprintf("urn:uuid:%s", id)
}
