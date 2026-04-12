package export

import (
	"encoding/xml"
	"fmt"
	"io"
)

// XMLEncoder encodes ExportData as XML using the standard library.
type XMLEncoder struct{}

// xmlExport is the top-level XML structure.
type xmlExport struct {
	XMLName xml.Name    `xml:"export"`
	Meta    xmlMeta     `xml:"meta"`
	Rows    xmlRows     `xml:"rows"`
}

// xmlMeta holds export metadata in XML.
type xmlMeta struct {
	Period       string `xml:"period"`
	KThreshold   int    `xml:"k_threshold"`
	Suppressed   int    `xml:"suppressed_groups"`
	TotalRecords int    `xml:"total_records"`
	GeneratedAt  string `xml:"generated_at"`
}

// xmlRows wraps a slice of XML rows.
type xmlRows struct {
	Row []xmlRow `xml:"row"`
}

// xmlRow represents a single row of data in XML.
type xmlRow struct {
	Fields []xmlField `xml:"field"`
}

// xmlField is a name-value pair within a row.
type xmlField struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

// Encode writes the export data as XML to w.
// Structure: <export><meta>...</meta><rows><row>...</row></rows></export>
func (e *XMLEncoder) Encode(w io.Writer, data ExportData) error {
	labelKeys, valueKeys := allColumns(data.Rows)

	rows := make([]xmlRow, 0, len(data.Rows))
	for _, r := range data.Rows {
		var fields []xmlField
		for _, k := range labelKeys {
			fields = append(fields, xmlField{Name: k, Value: r.Labels[k]})
		}
		for _, k := range valueKeys {
			fields = append(fields, xmlField{Name: k, Value: formatValue(r.Values[k])})
		}
		rows = append(rows, xmlRow{Fields: fields})
	}

	export := xmlExport{
		Meta: xmlMeta{
			Period:       data.Metadata.Period,
			KThreshold:   data.Metadata.KThreshold,
			Suppressed:   data.Metadata.Suppressed,
			TotalRecords: data.Metadata.TotalRecords,
			GeneratedAt:  data.Metadata.GeneratedAt.Format("2006-01-02T15:04:05Z"),
		},
		Rows: xmlRows{Row: rows},
	}

	if _, err := io.WriteString(w, xml.Header); err != nil {
		return fmt.Errorf("xml write header: %w", err)
	}

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(export); err != nil {
		return fmt.Errorf("xml encode: %w", err)
	}
	return nil
}

// ContentType returns the MIME type for XML.
func (e *XMLEncoder) ContentType() string {
	return "application/xml"
}

// FileExtension returns the file extension for XML.
func (e *XMLEncoder) FileExtension() string {
	return "xml"
}
