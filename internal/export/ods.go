package export

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"
)

// ODSEncoder encodes ExportData as an ODS (Open Document Spreadsheet) file.
// ODS is a ZIP archive containing XML files. We generate the minimal set of
// files required: META-INF/manifest.xml, content.xml, and mimetype.
type ODSEncoder struct{}

// Encode writes the export data as an ODS file to w.
func (e *ODSEncoder) Encode(w io.Writer, data ExportData) error {
	labelKeys, valueKeys := allColumns(data.Rows)
	allKeys := make([]string, 0, len(labelKeys)+len(valueKeys))
	allKeys = append(allKeys, labelKeys...)
	allKeys = append(allKeys, valueKeys...)

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// The mimetype file MUST be first and stored (not compressed) per ODF spec.
	mimetypeHeader := &zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store,
	}
	mw, err := zipWriter.CreateHeader(mimetypeHeader)
	if err != nil {
		return fmt.Errorf("ods create mimetype: %w", err)
	}
	if _, err := mw.Write([]byte("application/vnd.oasis.opendocument.spreadsheet")); err != nil {
		return fmt.Errorf("ods write mimetype: %w", err)
	}

	// manifest.xml
	manifestXML := `<?xml version="1.0" encoding="UTF-8"?>
<manifest:manifest xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0" manifest:version="1.2">
  <manifest:file-entry manifest:full-path="/" manifest:version="1.2" manifest:media-type="application/vnd.oasis.opendocument.spreadsheet"/>
  <manifest:file-entry manifest:full-path="content.xml" manifest:media-type="text/xml"/>
</manifest:manifest>`

	mfWriter, err := zipWriter.Create("META-INF/manifest.xml")
	if err != nil {
		return fmt.Errorf("ods create manifest: %w", err)
	}
	if _, err := mfWriter.Write([]byte(manifestXML)); err != nil {
		return fmt.Errorf("ods write manifest: %w", err)
	}

	// content.xml
	contentXML := buildODSContent(allKeys, labelKeys, valueKeys, data.Rows)
	cw, err := zipWriter.Create("content.xml")
	if err != nil {
		return fmt.Errorf("ods create content: %w", err)
	}
	if _, err := cw.Write([]byte(contentXML)); err != nil {
		return fmt.Errorf("ods write content: %w", err)
	}

	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("ods close zip: %w", err)
	}

	if _, err := w.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("ods write: %w", err)
	}
	return nil
}

// ContentType returns the MIME type for ODS.
func (e *ODSEncoder) ContentType() string {
	return "application/vnd.oasis.opendocument.spreadsheet"
}

// FileExtension returns the file extension for ODS.
func (e *ODSEncoder) FileExtension() string {
	return "ods"
}

// buildODSContent generates the content.xml for the ODS file.
func buildODSContent(allKeys, labelKeys, valueKeys []string, rows []ExportRow) string {
	var sb strings.Builder

	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString(`<office:document-content `)
	sb.WriteString(`xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" `)
	sb.WriteString(`xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0" `)
	sb.WriteString(`xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" `)
	sb.WriteString(`office:version="1.2">`)
	sb.WriteString(`<office:body><office:spreadsheet>`)
	sb.WriteString(`<table:table table:name="Export">`)

	// Header row.
	sb.WriteString(`<table:table-row>`)
	for _, key := range allKeys {
		sb.WriteString(`<table:table-cell><text:p>`)
		sb.WriteString(escapeXML(key))
		sb.WriteString(`</text:p></table:table-cell>`)
	}
	sb.WriteString(`</table:table-row>`)

	// Data rows.
	for _, row := range rows {
		sb.WriteString(`<table:table-row>`)
		for _, k := range labelKeys {
			sb.WriteString(`<table:table-cell><text:p>`)
			sb.WriteString(escapeXML(row.Labels[k]))
			sb.WriteString(`</text:p></table:table-cell>`)
		}
		for _, k := range valueKeys {
			sb.WriteString(`<table:table-cell><text:p>`)
			sb.WriteString(escapeXML(formatValue(row.Values[k])))
			sb.WriteString(`</text:p></table:table-cell>`)
		}
		sb.WriteString(`</table:table-row>`)
	}

	sb.WriteString(`</table:table>`)
	sb.WriteString(`</office:spreadsheet></office:body>`)
	sb.WriteString(`</office:document-content>`)

	return sb.String()
}

// escapeXML escapes special XML characters.
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
