// Package fhir provides FHIR R4 resource structs matching BR Core profiles
// for anonymized analytical data export. All structs are manually defined
// (no external FHIR library).
package fhir

// Bundle represents a FHIR R4 Bundle resource (type=collection).
type Bundle struct {
	ResourceType string        `json:"resourceType"`
	Type         string        `json:"type"`
	Timestamp    string        `json:"timestamp"`
	Total        int           `json:"total"`
	Entry        []BundleEntry `json:"entry"`
}

// BundleEntry is a single entry in a FHIR Bundle.
type BundleEntry struct {
	FullURL  string   `json:"fullUrl"`
	Resource Resource `json:"resource"`
}

// Resource is a generic FHIR resource. Each concrete type (Patient,
// Observation, Condition, Encounter) embeds common fields and adds
// type-specific ones.
type Resource struct {
	ResourceType string        `json:"resourceType"`
	ID           string        `json:"id,omitempty"`
	Meta         *Meta         `json:"meta,omitempty"`
	Identifier   []Identifier  `json:"identifier,omitempty"`
	Extension    []Extension   `json:"extension,omitempty"`
	Gender       string        `json:"gender,omitempty"`
	Address      []Address     `json:"address,omitempty"`
	Code         *CodeableConcept `json:"code,omitempty"`
	Subject      *Reference    `json:"subject,omitempty"`
	Status       string        `json:"status,omitempty"`
	Category     []CodeableConcept `json:"category,omitempty"`
	ValueInteger *int          `json:"valueInteger,omitempty"`
	EffectivePeriod *Period    `json:"effectivePeriod,omitempty"`
	ClinicalStatus  *CodeableConcept `json:"clinicalStatus,omitempty"`
	Class        *Coding       `json:"class,omitempty"`
	Type         []CodeableConcept `json:"type,omitempty"`
	Period       *Period       `json:"period,omitempty"`
}

// Meta contains resource metadata.
type Meta struct {
	Profile   []string `json:"profile,omitempty"`
	LastUpdated string `json:"lastUpdated,omitempty"`
}

// Identifier is a FHIR Identifier.
type Identifier struct {
	System string `json:"system"`
	Value  string `json:"value"`
}

// Extension is a FHIR Extension.
type Extension struct {
	URL         string `json:"url"`
	ValueString string `json:"valueString,omitempty"`
}

// Address is a FHIR Address.
type Address struct {
	District string `json:"district,omitempty"`
	State    string `json:"state,omitempty"`
}

// CodeableConcept is a FHIR CodeableConcept.
type CodeableConcept struct {
	Coding []Coding `json:"coding,omitempty"`
	Text   string   `json:"text,omitempty"`
}

// Coding is a FHIR Coding.
type Coding struct {
	System  string `json:"system,omitempty"`
	Code    string `json:"code,omitempty"`
	Display string `json:"display,omitempty"`
}

// Reference is a FHIR Reference.
type Reference struct {
	Reference string `json:"reference,omitempty"`
}

// Period is a FHIR Period.
type Period struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}
