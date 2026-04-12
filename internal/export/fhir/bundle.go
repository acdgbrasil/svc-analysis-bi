package fhir

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"fmt"
)

// NewBundle creates a FHIR R4 Bundle of type "collection".
func NewBundle(timestamp string, entries []BundleEntry) Bundle {
	return Bundle{
		ResourceType: "Bundle",
		Type:         "collection",
		Timestamp:    timestamp,
		Total:        len(entries),
		Entry:        entries,
	}
}

// GenerateID creates a deterministic ID from components using SHA-256 truncated to 16 hex chars.
// Components are joined with "|" delimiter to prevent hash collisions from concatenation ambiguity.
func GenerateID(components ...string) string {
	h := sha256.New()
	h.Write([]byte(strings.Join(components, "|")))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// NewBundleEntry creates a BundleEntry from a resource and its full URL.
func NewBundleEntry(fullURL string, resource Resource) BundleEntry {
	return BundleEntry{
		FullURL:  fullURL,
		Resource: resource,
	}
}

// PatientReference returns a FHIR reference string for a patient ID.
func PatientReference(patientID string) string {
	return fmt.Sprintf("Patient/%s", patientID)
}
