package fhir

import "fmt"

// FHIR system URIs for BR Core profiles.
const (
	SystemPatientHash = "http://acdgbrasil.com.br/fhir/patient-hash"
	SystemICD10       = "http://hl7.org/fhir/sid/icd-10"
	SystemACDG        = "http://acdgbrasil.com.br/fhir/valueset"
	SystemEncounter   = "http://terminology.hl7.org/CodeSystem/v3-ActCode"

	ProfilePatientBRCore    = "http://hl7.org/fhir/StructureDefinition/Patient"
	ProfileObservation      = "http://hl7.org/fhir/StructureDefinition/Observation"
	ProfileCondition        = "http://hl7.org/fhir/StructureDefinition/Condition"
	ProfileEncounter        = "http://hl7.org/fhir/StructureDefinition/Encounter"

	ExtensionAgeBand = "http://acdgbrasil.com.br/fhir/StructureDefinition/age-band"
)

// NewPatient creates a FHIR Patient resource from anonymized data.
// patientHash: SHA-256 hash identifier
// ageBand: 5-year age band (e.g. "0-4", "5-9")
// sex: MALE, FEMALE, or UNKNOWN
// mesoregion: IBGE mesoregion code
func NewPatient(id, patientHash, ageBand, sex, mesoregion string) Resource {
	gender := mapSexToFHIRGender(sex)

	return Resource{
		ResourceType: "Patient",
		ID:           id,
		Meta: &Meta{
			Profile: []string{ProfilePatientBRCore},
		},
		Identifier: []Identifier{
			{
				System: SystemPatientHash,
				Value:  patientHash,
			},
		},
		Extension: []Extension{
			{
				URL:         ExtensionAgeBand,
				ValueString: ageBand,
			},
		},
		Gender: gender,
		Address: []Address{
			{
				District: mesoregion,
			},
		},
	}
}

// mapSexToFHIRGender converts domain sex values to FHIR gender codes.
func mapSexToFHIRGender(sex string) string {
	switch sex {
	case "MALE":
		return "male"
	case "FEMALE":
		return "female"
	default:
		return "unknown"
	}
}

// PatientFullURL returns a URN for a patient resource.
func PatientFullURL(id string) string {
	return fmt.Sprintf("urn:uuid:%s", id)
}
