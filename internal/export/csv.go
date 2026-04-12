package export

import (
	"encoding/csv"
	"fmt"
	"io"
)

// CSVEncoder encodes ExportData as CSV using the standard library.
type CSVEncoder struct{}

// Encode writes the export data as CSV to w.
// Header row: label keys (sorted) + value keys (sorted).
// Data rows: one per ExportRow, values formatted as strings.
func (e *CSVEncoder) Encode(w io.Writer, data ExportData) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	labelKeys, valueKeys := allColumns(data.Rows)
	header := make([]string, 0, len(labelKeys)+len(valueKeys))
	header = append(header, labelKeys...)
	header = append(header, valueKeys...)

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("csv write header: %w", err)
	}

	for i, row := range data.Rows {
		record := make([]string, 0, len(header))
		for _, k := range labelKeys {
			record = append(record, row.Labels[k])
		}
		for _, k := range valueKeys {
			record = append(record, formatValue(row.Values[k]))
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("csv write row %d: %w", i, err)
		}
	}

	writer.Flush()
	return writer.Error()
}

// ContentType returns the MIME type for CSV.
func (e *CSVEncoder) ContentType() string {
	return "text/csv"
}

// FileExtension returns the file extension for CSV.
func (e *CSVEncoder) FileExtension() string {
	return "csv"
}
