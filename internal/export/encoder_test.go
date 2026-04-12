package export

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/acdgbrasil/svc-analysis-bi/internal/export/fhir"
)

// sampleData returns a consistent ExportData for testing.
func sampleData() ExportData {
	return ExportData{
		Dataset: "demographics",
		Rows: []ExportRow{
			{
				Labels: map[string]string{
					"age_band":   "0-4",
					"sex":        "MALE",
					"mesoregion": "3106",
				},
				Values: map[string]any{
					"count": 42,
					"rate":  0.15,
				},
			},
			{
				Labels: map[string]string{
					"age_band":   "5-9",
					"sex":        "FEMALE",
					"mesoregion": "3107",
				},
				Values: map[string]any{
					"count": 18,
					"rate":  0.08,
				},
			},
		},
		Metadata: ExportMetadata{
			Period:       "2025-01",
			KThreshold:   5,
			Suppressed:   2,
			TotalRecords: 60,
			GeneratedAt:  time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
}

// emptyData returns ExportData with no rows but valid metadata.
func emptyData() ExportData {
	return ExportData{
		Dataset:  "test",
		Rows:     []ExportRow{},
		Metadata: ExportMetadata{GeneratedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	expectedFormats := []string{"csv", "json", "xml", "parquet", "dbf", "dbc", "ods", "fhir"}
	for _, format := range expectedFormats {
		if _, ok := registry[format]; !ok {
			t.Errorf("registry missing format %q", format)
		}
	}

	if len(registry) != len(expectedFormats) {
		t.Errorf("registry has %d entries, want %d", len(registry), len(expectedFormats))
	}
}

func TestContentDisposition(t *testing.T) {
	got := ContentDisposition("demographics", "2025-01", "csv")
	want := `attachment; filename="acdg-demographics-2025-01.csv"`
	if got != want {
		t.Errorf("ContentDisposition = %q, want %q", got, want)
	}
}

// --- CSV Encoder Tests ---

func TestCSVEncoder_Encode(t *testing.T) {
	enc := &CSVEncoder{}
	var buf bytes.Buffer
	data := sampleData()

	if err := enc.Encode(&buf, data); err != nil {
		t.Fatalf("CSV Encode: %v", err)
	}

	// Parse back.
	reader := csv.NewReader(&buf)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("CSV ReadAll: %v", err)
	}

	// Header + 2 data rows.
	if len(records) != 3 {
		t.Fatalf("CSV records = %d, want 3", len(records))
	}

	header := records[0]
	// Label keys sorted + value keys sorted.
	wantHeader := []string{"age_band", "mesoregion", "sex", "count", "rate"}
	if len(header) != len(wantHeader) {
		t.Fatalf("CSV header cols = %d, want %d", len(header), len(wantHeader))
	}
	for i, h := range wantHeader {
		if header[i] != h {
			t.Errorf("CSV header[%d] = %q, want %q", i, header[i], h)
		}
	}

	// First data row.
	if records[1][0] != "0-4" {
		t.Errorf("CSV row1 age_band = %q, want %q", records[1][0], "0-4")
	}
}

func TestCSVEncoder_EmptyRows(t *testing.T) {
	enc := &CSVEncoder{}
	var buf bytes.Buffer

	if err := enc.Encode(&buf, emptyData()); err != nil {
		t.Fatalf("CSV Encode empty: %v", err)
	}

	// Should produce an empty header line only (no columns = no output).
	if buf.Len() != 0 {
		reader := csv.NewReader(&buf)
		records, _ := reader.ReadAll()
		// With no columns, we get one empty record (the header).
		if len(records) > 1 {
			t.Errorf("CSV empty should have at most 1 record, got %d", len(records))
		}
	}
}

func TestCSVEncoder_ContentType(t *testing.T) {
	enc := &CSVEncoder{}
	if enc.ContentType() != "text/csv" {
		t.Errorf("ContentType = %q, want %q", enc.ContentType(), "text/csv")
	}
	if enc.FileExtension() != "csv" {
		t.Errorf("FileExtension = %q, want %q", enc.FileExtension(), "csv")
	}
}

// --- JSON Encoder Tests ---

func TestJSONEncoder_Encode(t *testing.T) {
	enc := &JSONEncoder{}
	var buf bytes.Buffer
	data := sampleData()

	if err := enc.Encode(&buf, data); err != nil {
		t.Fatalf("JSON Encode: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON Unmarshal: %v", err)
	}

	// Verify top-level keys.
	if _, ok := result["data"]; !ok {
		t.Error("JSON missing 'data' key")
	}
	if _, ok := result["meta"]; !ok {
		t.Error("JSON missing 'meta' key")
	}

	// Verify data array.
	dataArr, ok := result["data"].([]any)
	if !ok {
		t.Fatal("JSON 'data' is not an array")
	}
	if len(dataArr) != 2 {
		t.Errorf("JSON data length = %d, want 2", len(dataArr))
	}

	// Verify first row has merged labels and values.
	firstRow, ok := dataArr[0].(map[string]any)
	if !ok {
		t.Fatal("JSON data[0] is not an object")
	}
	if firstRow["age_band"] != "0-4" {
		t.Errorf("JSON data[0].age_band = %v, want %q", firstRow["age_band"], "0-4")
	}

	// Verify metadata.
	meta, ok := result["meta"].(map[string]any)
	if !ok {
		t.Fatal("JSON 'meta' is not an object")
	}
	if meta["period"] != "2025-01" {
		t.Errorf("JSON meta.period = %v, want %q", meta["period"], "2025-01")
	}
	if meta["k_threshold"] != float64(5) {
		t.Errorf("JSON meta.k_threshold = %v, want 5", meta["k_threshold"])
	}
}

func TestJSONEncoder_ContentType(t *testing.T) {
	enc := &JSONEncoder{}
	if enc.ContentType() != "application/json" {
		t.Errorf("ContentType = %q", enc.ContentType())
	}
	if enc.FileExtension() != "json" {
		t.Errorf("FileExtension = %q", enc.FileExtension())
	}
}

// --- XML Encoder Tests ---

func TestXMLEncoder_Encode(t *testing.T) {
	enc := &XMLEncoder{}
	var buf bytes.Buffer
	data := sampleData()

	if err := enc.Encode(&buf, data); err != nil {
		t.Fatalf("XML Encode: %v", err)
	}

	// Verify it's valid XML by parsing.
	var result xmlExport
	if err := xml.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("XML Unmarshal: %v", err)
	}

	if result.Meta.Period != "2025-01" {
		t.Errorf("XML meta.period = %q, want %q", result.Meta.Period, "2025-01")
	}
	if result.Meta.KThreshold != 5 {
		t.Errorf("XML meta.k_threshold = %d, want 5", result.Meta.KThreshold)
	}
	if len(result.Rows.Row) != 2 {
		t.Errorf("XML rows = %d, want 2", len(result.Rows.Row))
	}

	// Verify first row has the right number of fields.
	if len(result.Rows.Row[0].Fields) != 5 {
		t.Errorf("XML row[0] fields = %d, want 5", len(result.Rows.Row[0].Fields))
	}
}

func TestXMLEncoder_ContentType(t *testing.T) {
	enc := &XMLEncoder{}
	if enc.ContentType() != "application/xml" {
		t.Errorf("ContentType = %q", enc.ContentType())
	}
	if enc.FileExtension() != "xml" {
		t.Errorf("FileExtension = %q", enc.FileExtension())
	}
}

// --- Parquet Encoder Tests ---

func TestParquetEncoder_Encode(t *testing.T) {
	enc := &ParquetEncoder{}
	var buf bytes.Buffer
	data := sampleData()

	if err := enc.Encode(&buf, data); err != nil {
		t.Fatalf("Parquet Encode: %v", err)
	}

	b := buf.Bytes()

	// Verify PAR1 magic at start.
	if len(b) < 4 || string(b[:4]) != "PAR1" {
		t.Error("Parquet: missing PAR1 magic at start")
	}

	// Verify PAR1 magic at end.
	if len(b) < 4 || string(b[len(b)-4:]) != "PAR1" {
		t.Error("Parquet: missing PAR1 magic at end")
	}

	// Minimum size sanity check.
	if len(b) < 20 {
		t.Errorf("Parquet: file too small (%d bytes)", len(b))
	}
}

func TestParquetEncoder_NoColumns(t *testing.T) {
	enc := &ParquetEncoder{}
	var buf bytes.Buffer
	data := ExportData{
		Dataset: "test",
		Rows:    []ExportRow{{Labels: map[string]string{}, Values: map[string]any{}}},
	}

	err := enc.Encode(&buf, data)
	if err == nil {
		t.Error("Parquet: expected error for no columns")
	}
}

func TestParquetEncoder_ContentType(t *testing.T) {
	enc := &ParquetEncoder{}
	if enc.ContentType() != "application/vnd.apache.parquet" {
		t.Errorf("ContentType = %q", enc.ContentType())
	}
	if enc.FileExtension() != "parquet" {
		t.Errorf("FileExtension = %q", enc.FileExtension())
	}
}

// --- DBF Encoder Tests ---

func TestDBFEncoder_Encode(t *testing.T) {
	enc := &DBFEncoder{}
	var buf bytes.Buffer
	data := sampleData()

	if err := enc.Encode(&buf, data); err != nil {
		t.Fatalf("DBF Encode: %v", err)
	}

	b := buf.Bytes()

	// Verify dBASE III version byte.
	if b[0] != dbfVersion {
		t.Errorf("DBF version byte = 0x%02X, want 0x%02X", b[0], dbfVersion)
	}

	// Verify record count (bytes 4-7, LE uint32).
	numRecords := uint32(b[4]) | uint32(b[5])<<8 | uint32(b[6])<<16 | uint32(b[7])<<24
	if numRecords != 2 {
		t.Errorf("DBF record count = %d, want 2", numRecords)
	}

	// Verify file ends with 0x1A terminator.
	if b[len(b)-1] != dbfRecordTerminator {
		t.Errorf("DBF last byte = 0x%02X, want 0x%02X", b[len(b)-1], dbfRecordTerminator)
	}
}

func TestDBFEncoder_FieldNameTruncation(t *testing.T) {
	name := truncateFieldName("very_long_field_name")
	if len(name) > 10 {
		t.Errorf("truncateFieldName: len = %d, want <= 10", len(name))
	}
	if name != "very_long_" {
		t.Errorf("truncateFieldName = %q, want %q", name, "very_long_")
	}
}

func TestDBFEncoder_ContentType(t *testing.T) {
	enc := &DBFEncoder{}
	if enc.ContentType() != "application/x-dbf" {
		t.Errorf("ContentType = %q", enc.ContentType())
	}
	if enc.FileExtension() != "dbf" {
		t.Errorf("FileExtension = %q", enc.FileExtension())
	}
}

// --- DBC Encoder Tests ---

func TestDBCEncoder_Encode(t *testing.T) {
	enc := &DBCEncoder{}
	var buf bytes.Buffer
	data := sampleData()

	if err := enc.Encode(&buf, data); err != nil {
		t.Fatalf("DBC Encode: %v", err)
	}

	b := buf.Bytes()

	// Verify header: 8 bytes (original size + compressed size).
	if len(b) < dbcHeaderSize {
		t.Fatalf("DBC: file too small (%d bytes)", len(b))
	}

	origSize := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	compSize := uint32(b[4]) | uint32(b[5])<<8 | uint32(b[6])<<16 | uint32(b[7])<<24

	if origSize == 0 {
		t.Error("DBC: original size is 0")
	}
	if compSize == 0 {
		t.Error("DBC: compressed size is 0")
	}

	// Verify compressed data can be decompressed.
	compressedData := b[dbcHeaderSize:]
	if uint32(len(compressedData)) != compSize {
		t.Errorf("DBC: compressed data len = %d, header says %d", len(compressedData), compSize)
	}

	reader := flate.NewReader(bytes.NewReader(compressedData))
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("DBC: decompression failed: %v", err)
	}

	if uint32(len(decompressed)) != origSize {
		t.Errorf("DBC: decompressed len = %d, header says %d", len(decompressed), origSize)
	}

	// Verify decompressed data is valid DBF (starts with version byte).
	if decompressed[0] != dbfVersion {
		t.Errorf("DBC: decompressed DBF version = 0x%02X, want 0x%02X", decompressed[0], dbfVersion)
	}
}

func TestDBCEncoder_ContentType(t *testing.T) {
	enc := &DBCEncoder{}
	if enc.ContentType() != "application/x-dbc" {
		t.Errorf("ContentType = %q", enc.ContentType())
	}
	if enc.FileExtension() != "dbc" {
		t.Errorf("FileExtension = %q", enc.FileExtension())
	}
}

// --- ODS Encoder Tests ---

func TestODSEncoder_Encode(t *testing.T) {
	enc := &ODSEncoder{}
	var buf bytes.Buffer
	data := sampleData()

	if err := enc.Encode(&buf, data); err != nil {
		t.Fatalf("ODS Encode: %v", err)
	}

	// Verify it's a valid ZIP file.
	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("ODS: not a valid ZIP: %v", err)
	}

	// Check required files.
	fileNames := make(map[string]bool)
	for _, f := range reader.File {
		fileNames[f.Name] = true
	}

	if !fileNames["mimetype"] {
		t.Error("ODS: missing mimetype file")
	}
	if !fileNames["content.xml"] {
		t.Error("ODS: missing content.xml")
	}
	if !fileNames["META-INF/manifest.xml"] {
		t.Error("ODS: missing META-INF/manifest.xml")
	}

	// Verify mimetype is stored (not compressed) and has correct content.
	for _, f := range reader.File {
		if f.Name == "mimetype" {
			if f.Method != zip.Store {
				t.Error("ODS: mimetype should be stored, not compressed")
			}
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("ODS: open mimetype: %v", err)
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("ODS: read mimetype: %v", err)
			}
			if string(content) != "application/vnd.oasis.opendocument.spreadsheet" {
				t.Errorf("ODS: mimetype = %q", string(content))
			}
		}
	}

	// Verify content.xml contains table data.
	for _, f := range reader.File {
		if f.Name == "content.xml" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("ODS: open content.xml: %v", err)
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("ODS: read content.xml: %v", err)
			}
			s := string(content)
			if !strings.Contains(s, "table:table") {
				t.Error("ODS: content.xml missing table element")
			}
			if !strings.Contains(s, "0-4") {
				t.Error("ODS: content.xml missing data value '0-4'")
			}
		}
	}
}

func TestODSEncoder_ContentType(t *testing.T) {
	enc := &ODSEncoder{}
	if enc.ContentType() != "application/vnd.oasis.opendocument.spreadsheet" {
		t.Errorf("ContentType = %q", enc.ContentType())
	}
	if enc.FileExtension() != "ods" {
		t.Errorf("FileExtension = %q", enc.FileExtension())
	}
}

// --- FHIR Encoder Tests ---

func TestFHIREncoder_Encode_Observations(t *testing.T) {
	enc := &FHIREncoder{}
	var buf bytes.Buffer

	data := ExportData{
		Dataset: "demographics",
		Rows: []ExportRow{
			{
				Labels: map[string]string{
					"indicator_code": "IND-001",
					"indicator_name": "Patient Count",
					"patient_ref":    "Patient/abc123",
				},
				Values: map[string]any{
					"count": 42,
				},
			},
		},
		Metadata: ExportMetadata{
			Period:       "2025-01",
			KThreshold:   5,
			TotalRecords: 1,
			GeneratedAt:  time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}

	if err := enc.Encode(&buf, data); err != nil {
		t.Fatalf("FHIR Encode: %v", err)
	}

	var bundle fhir.Bundle
	if err := json.Unmarshal(buf.Bytes(), &bundle); err != nil {
		t.Fatalf("FHIR Unmarshal: %v", err)
	}

	if bundle.ResourceType != "Bundle" {
		t.Errorf("FHIR resourceType = %q, want %q", bundle.ResourceType, "Bundle")
	}
	if bundle.Type != "collection" {
		t.Errorf("FHIR type = %q, want %q", bundle.Type, "collection")
	}
	if bundle.Total != 1 {
		t.Errorf("FHIR total = %d, want 1", bundle.Total)
	}
	if len(bundle.Entry) != 1 {
		t.Fatalf("FHIR entries = %d, want 1", len(bundle.Entry))
	}

	resource := bundle.Entry[0].Resource
	if resource.ResourceType != "Observation" {
		t.Errorf("FHIR entry resourceType = %q, want %q", resource.ResourceType, "Observation")
	}
	if resource.Status != "final" {
		t.Errorf("FHIR entry status = %q, want %q", resource.Status, "final")
	}
}

func TestFHIREncoder_Encode_Patient(t *testing.T) {
	enc := &FHIREncoder{}
	var buf bytes.Buffer

	data := ExportData{
		Dataset: "demographics",
		Rows: []ExportRow{
			{
				Labels: map[string]string{
					"patient_hash": "abc123hash",
					"age_band":     "5-9",
					"sex":          "FEMALE",
					"mesoregion":   "3106",
				},
				Values: map[string]any{},
			},
		},
		Metadata: ExportMetadata{
			Period:      "2025-01",
			GeneratedAt: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}

	if err := enc.Encode(&buf, data); err != nil {
		t.Fatalf("FHIR Encode Patient: %v", err)
	}

	var bundle fhir.Bundle
	if err := json.Unmarshal(buf.Bytes(), &bundle); err != nil {
		t.Fatalf("FHIR Unmarshal: %v", err)
	}

	if len(bundle.Entry) != 1 {
		t.Fatalf("FHIR entries = %d, want 1", len(bundle.Entry))
	}

	resource := bundle.Entry[0].Resource
	if resource.ResourceType != "Patient" {
		t.Errorf("resourceType = %q, want Patient", resource.ResourceType)
	}
	if resource.Gender != "female" {
		t.Errorf("gender = %q, want female", resource.Gender)
	}
	if len(resource.Identifier) == 0 || resource.Identifier[0].Value != "abc123hash" {
		t.Error("FHIR Patient missing identifier")
	}
}

func TestFHIREncoder_Encode_Condition(t *testing.T) {
	enc := &FHIREncoder{}
	var buf bytes.Buffer

	data := ExportData{
		Dataset: "epidemiological",
		Rows: []ExportRow{
			{
				Labels: map[string]string{
					"icd_code":       "Q90.0",
					"condition_name": "Down syndrome",
					"patient_ref":    "Patient/xyz",
				},
				Values: map[string]any{},
			},
		},
		Metadata: ExportMetadata{
			Period:      "2025-01",
			GeneratedAt: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}

	if err := enc.Encode(&buf, data); err != nil {
		t.Fatalf("FHIR Encode Condition: %v", err)
	}

	var bundle fhir.Bundle
	if err := json.Unmarshal(buf.Bytes(), &bundle); err != nil {
		t.Fatalf("FHIR Unmarshal: %v", err)
	}

	resource := bundle.Entry[0].Resource
	if resource.ResourceType != "Condition" {
		t.Errorf("resourceType = %q, want Condition", resource.ResourceType)
	}
	if resource.Code == nil || len(resource.Code.Coding) == 0 {
		t.Fatal("Condition missing code")
	}
	if resource.Code.Coding[0].System != "http://hl7.org/fhir/sid/icd-10" {
		t.Errorf("Condition code system = %q", resource.Code.Coding[0].System)
	}
	if resource.Code.Coding[0].Code != "Q90.0" {
		t.Errorf("Condition code = %q, want Q90.0", resource.Code.Coding[0].Code)
	}
}

func TestFHIREncoder_Encode_Encounter(t *testing.T) {
	enc := &FHIREncoder{}
	var buf bytes.Buffer

	data := ExportData{
		Dataset: "care",
		Rows: []ExportRow{
			{
				Labels: map[string]string{
					"encounter_type": "Initial Assessment",
					"class":          "AMB",
					"patient_ref":    "Patient/xyz",
				},
				Values: map[string]any{},
			},
		},
		Metadata: ExportMetadata{
			Period:      "2025-01",
			GeneratedAt: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}

	if err := enc.Encode(&buf, data); err != nil {
		t.Fatalf("FHIR Encode Encounter: %v", err)
	}

	var bundle fhir.Bundle
	if err := json.Unmarshal(buf.Bytes(), &bundle); err != nil {
		t.Fatalf("FHIR Unmarshal: %v", err)
	}

	resource := bundle.Entry[0].Resource
	if resource.ResourceType != "Encounter" {
		t.Errorf("resourceType = %q, want Encounter", resource.ResourceType)
	}
	if resource.Status != "finished" {
		t.Errorf("status = %q, want finished", resource.Status)
	}
}

func TestFHIREncoder_ContentType(t *testing.T) {
	enc := &FHIREncoder{}
	if enc.ContentType() != "application/fhir+json" {
		t.Errorf("ContentType = %q", enc.ContentType())
	}
	if enc.FileExtension() != "json" {
		t.Errorf("FileExtension = %q", enc.FileExtension())
	}
}

// --- Helper Function Tests ---

func TestOrderedKeys(t *testing.T) {
	m := map[string]string{"c": "3", "a": "1", "b": "2"}
	keys := orderedKeys(m)
	want := []string{"a", "b", "c"}
	if len(keys) != len(want) {
		t.Fatalf("orderedKeys len = %d, want %d", len(keys), len(want))
	}
	for i, k := range want {
		if keys[i] != k {
			t.Errorf("orderedKeys[%d] = %q, want %q", i, keys[i], k)
		}
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		input any
		want  string
	}{
		{42, "42"},
		{3.14, "3.14"},
		{"hello", "hello"},
		{true, "true"},
		{nil, ""},
	}
	for _, tt := range tests {
		got := formatValue(tt.input)
		if got != tt.want {
			t.Errorf("formatValue(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- Interface Compliance Tests ---

func TestEncoderInterfaceCompliance(t *testing.T) {
	encoders := []Encoder{
		&CSVEncoder{},
		&JSONEncoder{},
		&XMLEncoder{},
		&ParquetEncoder{},
		&DBFEncoder{},
		&DBCEncoder{},
		&ODSEncoder{},
		&FHIREncoder{},
	}

	data := sampleData()

	for _, enc := range encoders {
		t.Run(enc.FileExtension(), func(t *testing.T) {
			var buf bytes.Buffer
			if err := enc.Encode(&buf, data); err != nil {
				t.Fatalf("Encode failed: %v", err)
			}
			if buf.Len() == 0 {
				t.Error("Encode produced empty output")
			}
			if enc.ContentType() == "" {
				t.Error("ContentType is empty")
			}
			if enc.FileExtension() == "" {
				t.Error("FileExtension is empty")
			}
		})
	}
}

func TestEscapeXML(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"a&b", "a&amp;b"},
		{"<tag>", "&lt;tag&gt;"},
		{`"quoted"`, "&quot;quoted&quot;"},
		{"it's", "it&apos;s"},
	}
	for _, tt := range tests {
		got := escapeXML(tt.input)
		if got != tt.want {
			t.Errorf("escapeXML(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestToInt(t *testing.T) {
	tests := []struct {
		input any
		want  int
	}{
		{42, 42},
		{int64(100), 100},
		{float64(3.7), 3},
		{"15", 15},
		{"not-a-number", 0},
		{nil, 0},
		{true, 0},
	}
	for _, tt := range tests {
		got := toInt(tt.input)
		if got != tt.want {
			t.Errorf("toInt(%v) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
