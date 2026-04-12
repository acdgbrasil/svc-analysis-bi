package export

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// JSONEncoder encodes ExportData as JSON using the standard library.
type JSONEncoder struct{}

// jsonEnvelope is the top-level JSON output structure.
type jsonEnvelope struct {
	Data []jsonRow    `json:"data"`
	Meta jsonMetadata `json:"meta"`
}

// jsonRow is a flattened row combining labels and values.
type jsonRow map[string]any

// jsonMetadata mirrors ExportMetadata for JSON serialization.
type jsonMetadata struct {
	Period       string `json:"period"`
	KThreshold   int    `json:"k_threshold"`
	Suppressed   int    `json:"suppressed_groups"`
	TotalRecords int    `json:"total_records"`
	GeneratedAt  string `json:"generated_at"`
}

// Encode writes the export data as JSON to w.
// Output: {"data": [...], "meta": {...}}
func (e *JSONEncoder) Encode(w io.Writer, data ExportData) error {
	rows := make([]jsonRow, 0, len(data.Rows))
	for _, r := range data.Rows {
		row := make(jsonRow, len(r.Labels)+len(r.Values))
		for k, v := range r.Labels {
			row[k] = v
		}
		for k, v := range r.Values {
			row[k] = v
		}
		rows = append(rows, row)
	}

	envelope := jsonEnvelope{
		Data: rows,
		Meta: jsonMetadata{
			Period:       data.Metadata.Period,
			KThreshold:   data.Metadata.KThreshold,
			Suppressed:   data.Metadata.Suppressed,
			TotalRecords: data.Metadata.TotalRecords,
			GeneratedAt:  data.Metadata.GeneratedAt.UTC().Format(time.RFC3339),
		},
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(envelope); err != nil {
		return fmt.Errorf("json encode: %w", err)
	}
	return nil
}

// ContentType returns the MIME type for JSON.
func (e *JSONEncoder) ContentType() string {
	return "application/json"
}

// FileExtension returns the file extension for JSON.
func (e *JSONEncoder) FileExtension() string {
	return "json"
}
