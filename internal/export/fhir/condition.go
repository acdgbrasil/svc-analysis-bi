package fhir

import "fmt"

// NewCondition creates a FHIR Condition resource from diagnosis data.
// icdCode: ICD-10 code (e.g. "Q90.0")
// display: human-readable condition name
// patientRef: reference to the Patient resource
func NewCondition(id, icdCode, display, patientRef string) Resource {
	return Resource{
		ResourceType: "Condition",
		ID:           id,
		Meta: &Meta{
			Profile: []string{ProfileCondition},
		},
		ClinicalStatus: &CodeableConcept{
			Coding: []Coding{
				{
					System:  "http://terminology.hl7.org/CodeSystem/condition-clinical",
					Code:    "active",
					Display: "Active",
				},
			},
		},
		Code: &CodeableConcept{
			Coding: []Coding{
				{
					System:  SystemICD10,
					Code:    icdCode,
					Display: display,
				},
			},
		},
		Subject: &Reference{
			Reference: patientRef,
		},
	}
}

// ConditionFullURL returns a URN for a condition resource.
func ConditionFullURL(id string) string {
	return fmt.Sprintf("urn:uuid:%s", id)
}
