// Package export provides 8 format-specific encoders for bulk data export.
// Each encoder implements the Encoder interface, writing ExportData to an io.Writer.
// Encoders are pure functions with no side effects beyond writing.
package export

import (
	"fmt"
	"io"
	"sort"
	"time"
)

// ExportData is the input to all encoders -- rows of indicator data.
type ExportData struct {
	Dataset  string
	Rows     []ExportRow
	Metadata ExportMetadata
}

// ExportRow represents a single row of export data with string labels and
// typed values.
type ExportRow struct {
	Labels map[string]string
	Values map[string]any // int, float64, string, bool
}

// ExportMetadata contains information about the exported dataset.
type ExportMetadata struct {
	Period       string
	KThreshold   int
	Suppressed   int
	TotalRecords int
	GeneratedAt  time.Time
}

// Encoder writes ExportData to a writer in a specific format.
type Encoder interface {
	Encode(w io.Writer, data ExportData) error
	ContentType() string
	FileExtension() string
}

// NewRegistry returns a map of format names to their corresponding encoders.
func NewRegistry() map[string]Encoder {
	return map[string]Encoder{
		"csv":     &CSVEncoder{},
		"json":    &JSONEncoder{},
		"xml":     &XMLEncoder{},
		"parquet": &ParquetEncoder{},
		"dbf":     &DBFEncoder{},
		"dbc":     &DBCEncoder{},
		"ods":     &ODSEncoder{},
		"fhir":    &FHIREncoder{},
	}
}

// ContentDisposition returns a properly formatted Content-Disposition header
// value for file download.
func ContentDisposition(dataset, period, ext string) string {
	return fmt.Sprintf(`attachment; filename="acdg-%s-%s.%s"`, dataset, period, ext)
}

// orderedKeys returns sorted keys from a string map for deterministic output.
func orderedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// orderedAnyKeys returns sorted keys from a map[string]any for deterministic output.
func orderedAnyKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// allColumns scans all rows and returns the union of all label and value keys,
// label keys first (sorted), then value keys (sorted).
func allColumns(rows []ExportRow) (labelKeys, valueKeys []string) {
	labelSet := make(map[string]struct{})
	valueSet := make(map[string]struct{})
	for _, r := range rows {
		for k := range r.Labels {
			labelSet[k] = struct{}{}
		}
		for k := range r.Values {
			valueSet[k] = struct{}{}
		}
	}
	labelKeys = make([]string, 0, len(labelSet))
	for k := range labelSet {
		labelKeys = append(labelKeys, k)
	}
	sort.Strings(labelKeys)

	valueKeys = make([]string, 0, len(valueSet))
	for k := range valueSet {
		valueKeys = append(valueKeys, k)
	}
	sort.Strings(valueKeys)
	return labelKeys, valueKeys
}

// formatValue converts any value to its string representation.
func formatValue(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
