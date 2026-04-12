package export

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/acdgbrasil/svc-analysis-bi/internal/export/fhir"
)

// FHIREncoder encodes ExportData as a FHIR R4 Bundle (type=collection) in JSON.
// It maps export rows to FHIR resources based on label conventions:
//   - Rows with "patient_hash" label: Patient resources
//   - Rows with "icd_code" label: Condition resources
//   - Rows with "encounter_type" label: Encounter resources
//   - All other rows: Observation resources (indicator data)
type FHIREncoder struct{}

// Encode writes the export data as a FHIR Bundle to w.
func (e *FHIREncoder) Encode(w io.Writer, data ExportData) error {
	timestamp := data.Metadata.GeneratedAt.Format("2006-01-02T15:04:05Z")
	entries := make([]fhir.BundleEntry, 0, len(data.Rows))

	for i, row := range data.Rows {
		entry, err := rowToEntry(row, i, data.Metadata.Period)
		if err != nil {
			return fmt.Errorf("fhir encode row %d: %w", i, err)
		}
		entries = append(entries, entry)
	}

	bundle := fhir.NewBundle(timestamp, entries)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(bundle); err != nil {
		return fmt.Errorf("fhir json encode: %w", err)
	}
	return nil
}

// ContentType returns the MIME type for FHIR JSON.
func (e *FHIREncoder) ContentType() string {
	return "application/fhir+json"
}

// FileExtension returns the file extension for FHIR JSON.
func (e *FHIREncoder) FileExtension() string {
	return "json"
}

// rowToEntry converts an ExportRow to a FHIR BundleEntry based on its labels.
func rowToEntry(row ExportRow, index int, period string) (fhir.BundleEntry, error) {
	idSuffix := strconv.Itoa(index)

	// Patient: rows with patient_hash label.
	if hash, ok := row.Labels["patient_hash"]; ok {
		id := fhir.GenerateID("patient", idSuffix)
		ageBand := row.Labels["age_band"]
		sex := row.Labels["sex"]
		mesoregion := row.Labels["mesoregion"]
		resource := fhir.NewPatient(id, hash, ageBand, sex, mesoregion)
		return fhir.NewBundleEntry(fhir.PatientFullURL(id), resource), nil
	}

	// Condition: rows with icd_code label.
	if icdCode, ok := row.Labels["icd_code"]; ok {
		id := fhir.GenerateID("condition", idSuffix)
		display := row.Labels["condition_name"]
		patientRef := row.Labels["patient_ref"]
		if patientRef == "" {
			patientRef = "Patient/unknown"
		}
		resource := fhir.NewCondition(id, icdCode, display, patientRef)
		return fhir.NewBundleEntry(fhir.ConditionFullURL(id), resource), nil
	}

	// Encounter: rows with encounter_type label.
	if encType, ok := row.Labels["encounter_type"]; ok {
		id := fhir.GenerateID("encounter", idSuffix)
		classCode := row.Labels["class"]
		if classCode == "" {
			classCode = "AMB"
		}
		patientRef := row.Labels["patient_ref"]
		if patientRef == "" {
			patientRef = "Patient/unknown"
		}
		resource := fhir.NewEncounter(id, encType, classCode, period, period, patientRef)
		return fhir.NewBundleEntry(fhir.EncounterFullURL(id), resource), nil
	}

	// Default: Observation (indicator data).
	id := fhir.GenerateID("observation", idSuffix)
	code := row.Labels["indicator_code"]
	if code == "" {
		code = "unknown"
	}
	display := row.Labels["indicator_name"]
	patientRef := row.Labels["patient_ref"]
	if patientRef == "" {
		patientRef = "Patient/unknown"
	}

	value := 0
	if v, ok := row.Values["count"]; ok {
		value = toInt(v)
	} else if v, ok := row.Values["value"]; ok {
		value = toInt(v)
	}

	resource := fhir.NewObservation(id, code, display, value, period, period, patientRef)
	return fhir.NewBundleEntry(fhir.ObservationFullURL(id), resource), nil
}

// toInt converts an any value to int, best-effort.
func toInt(v any) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		n, _ := strconv.Atoi(val)
		return n
	default:
		return 0
	}
}
