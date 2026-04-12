package fhir

import "fmt"

// NewObservation creates a FHIR Observation resource from indicator data.
// code: indicator code from custom ValueSet
// display: human-readable indicator name
// value: integer indicator value
// periodStart/periodEnd: observation period
// patientRef: reference to the Patient resource (e.g. "Patient/xxx")
func NewObservation(id, code, display string, value int, periodStart, periodEnd, patientRef string) Resource {
	return Resource{
		ResourceType: "Observation",
		ID:           id,
		Meta: &Meta{
			Profile: []string{ProfileObservation},
		},
		Status: "final",
		Code: &CodeableConcept{
			Coding: []Coding{
				{
					System:  SystemACDG,
					Code:    code,
					Display: display,
				},
			},
		},
		Subject: &Reference{
			Reference: patientRef,
		},
		ValueInteger: &value,
		EffectivePeriod: &Period{
			Start: periodStart,
			End:   periodEnd,
		},
	}
}

// ObservationFullURL returns a URN for an observation resource.
func ObservationFullURL(id string) string {
	return fmt.Sprintf("urn:uuid:%s", id)
}
