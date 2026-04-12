package export

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

// DBFEncoder encodes ExportData in dBASE III format for DataSUS TABWIN
// compatibility. Field names are truncated to 10 characters per dBASE spec.
type DBFEncoder struct{}

// dBASE III constants.
const (
	dbfVersion     = 0x03 // dBASE III without memo
	dbfHeaderTerminator = 0x0D
	dbfRecordTerminator = 0x1A
	dbfFieldCharType    = 'C'
	dbfMaxFieldName     = 11 // 10 chars + null terminator
	dbfFieldDescSize    = 32
	dbfHeaderBaseSize   = 32
	dbfMaxFieldLen      = 254
)

// Encode writes the export data as a dBASE III file to w.
func (e *DBFEncoder) Encode(w io.Writer, data ExportData) error {
	labelKeys, valueKeys := allColumns(data.Rows)
	allKeys := make([]string, 0, len(labelKeys)+len(valueKeys))
	allKeys = append(allKeys, labelKeys...)
	allKeys = append(allKeys, valueKeys...)

	if len(allKeys) == 0 {
		return fmt.Errorf("dbf encode: no columns")
	}

	// Calculate field widths: scan all data to find max width per column.
	fieldWidths := make([]int, len(allKeys))
	for i := range fieldWidths {
		fieldWidths[i] = len(truncateFieldName(allKeys[i])) // minimum: header name length
		if fieldWidths[i] < 1 {
			fieldWidths[i] = 1
		}
	}

	for _, row := range data.Rows {
		colIdx := 0
		for _, k := range labelKeys {
			w := len(row.Labels[k])
			if w > fieldWidths[colIdx] {
				fieldWidths[colIdx] = w
			}
			colIdx++
		}
		for _, k := range valueKeys {
			w := len(formatValue(row.Values[k]))
			if w > fieldWidths[colIdx] {
				fieldWidths[colIdx] = w
			}
			colIdx++
		}
	}

	// Cap field widths at 254.
	for i := range fieldWidths {
		if fieldWidths[i] > dbfMaxFieldLen {
			fieldWidths[i] = dbfMaxFieldLen
		}
	}

	// Calculate record size: 1 (deletion flag) + sum of field widths.
	recordSize := 1
	for _, fw := range fieldWidths {
		recordSize += fw
	}

	numRecords := len(data.Rows)
	numFields := len(allKeys)
	headerSize := dbfHeaderBaseSize + (numFields * dbfFieldDescSize) + 1 // +1 for header terminator

	var buf bytes.Buffer

	// DBF Header (32 bytes).
	now := time.Now()
	buf.WriteByte(dbfVersion)
	buf.WriteByte(byte(now.Year() - 1900))
	buf.WriteByte(byte(now.Month()))
	buf.WriteByte(byte(now.Day()))

	// Number of records (4 bytes LE).
	recBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(recBytes, uint32(numRecords))
	buf.Write(recBytes)

	// Header size (2 bytes LE).
	hdrBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(hdrBytes, uint16(headerSize))
	buf.Write(hdrBytes)

	// Record size (2 bytes LE).
	recSizeBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(recSizeBytes, uint16(recordSize))
	buf.Write(recSizeBytes)

	// Reserved (20 bytes).
	buf.Write(make([]byte, 20))

	// Field descriptors (32 bytes each).
	for i, key := range allKeys {
		name := truncateFieldName(key)
		nameBytes := make([]byte, dbfMaxFieldName)
		copy(nameBytes, []byte(name))
		buf.Write(nameBytes)

		buf.WriteByte(byte(dbfFieldCharType)) // field type: Character

		buf.Write(make([]byte, 4)) // reserved

		buf.WriteByte(byte(fieldWidths[i])) // field length
		buf.WriteByte(0)                    // decimal count

		buf.Write(make([]byte, 14)) // reserved
	}

	buf.WriteByte(dbfHeaderTerminator)

	// Records.
	for _, row := range data.Rows {
		buf.WriteByte(' ') // deletion flag: space = not deleted

		colIdx := 0
		for _, k := range labelKeys {
			val := row.Labels[k]
			writeDBFField(&buf, val, fieldWidths[colIdx])
			colIdx++
		}
		for _, k := range valueKeys {
			val := formatValue(row.Values[k])
			writeDBFField(&buf, val, fieldWidths[colIdx])
			colIdx++
		}
	}

	buf.WriteByte(dbfRecordTerminator)

	if _, err := w.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("dbf write: %w", err)
	}
	return nil
}

// ContentType returns the MIME type for DBF.
func (e *DBFEncoder) ContentType() string {
	return "application/x-dbf"
}

// FileExtension returns the file extension for DBF.
func (e *DBFEncoder) FileExtension() string {
	return "dbf"
}

// truncateFieldName truncates a field name to 10 characters (dBASE III max).
func truncateFieldName(name string) string {
	if len(name) > 10 {
		return name[:10]
	}
	return name
}

// writeDBFField writes a fixed-width, space-padded character field.
func writeDBFField(buf *bytes.Buffer, val string, width int) {
	if len(val) > width {
		val = val[:width]
	}
	buf.WriteString(val)
	// Pad with spaces.
	for i := len(val); i < width; i++ {
		buf.WriteByte(' ')
	}
}

// EncodeDBFToBytes is a helper that encodes ExportData to DBF bytes.
// Used by the DBC encoder to get the raw DBF content before compression.
func EncodeDBFToBytes(data ExportData) ([]byte, error) {
	var buf bytes.Buffer
	enc := &DBFEncoder{}
	if err := enc.Encode(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
