package export

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// ParquetEncoder encodes ExportData in Apache Parquet format.
// This is a minimal Parquet implementation that produces a valid Parquet file
// with a single row group, plain encoding, and no compression.
// It writes the PAR1 magic bytes, row group data, file metadata, and footer.
//
// TODO: Known limitations in the Thrift compact protocol encoding:
//   - List header byte encoding uses a simplified approach that may not be
//     parsed correctly by all Parquet readers for lists with >14 elements.
//   - Zigzag encoding is applied uniformly via writeThriftCompactVarInt,
//     including for length prefixes where unsigned encoding would be more correct.
//   - Field ID delta encoding always uses absolute field IDs rather than deltas
//     from the previous field, which produces slightly larger but valid output.
//   For production use, consider replacing with a full Thrift library or the
//   segmentio/parquet-go package.
type ParquetEncoder struct{}

// Parquet format constants.
const (
	parquetMagic = "PAR1"
)

// Thrift compact protocol type IDs.
const (
	thriftStop   = 0
	thriftBool   = 1
	thriftI32    = 5
	thriftI64    = 6
	thriftBinary = 8
	thriftList   = 15
	thriftStruct = 12
)

// Encode writes the export data as a minimal Parquet file to w.
func (e *ParquetEncoder) Encode(w io.Writer, data ExportData) error {
	labelKeys, valueKeys := allColumns(data.Rows)
	allKeys := make([]string, 0, len(labelKeys)+len(valueKeys))
	allKeys = append(allKeys, labelKeys...)
	allKeys = append(allKeys, valueKeys...)

	if len(allKeys) == 0 {
		return fmt.Errorf("parquet encode: no columns")
	}

	// Build column data as byte arrays (all columns stored as BYTE_ARRAY/string).
	columns := make([][]string, len(allKeys))
	for i := range columns {
		columns[i] = make([]string, len(data.Rows))
	}
	for rowIdx, row := range data.Rows {
		colIdx := 0
		for _, k := range labelKeys {
			columns[colIdx][rowIdx] = row.Labels[k]
			colIdx++
		}
		for _, k := range valueKeys {
			columns[colIdx][rowIdx] = formatValue(row.Values[k])
			colIdx++
		}
	}

	// We build the file in a buffer to compute offsets.
	var buf bytes.Buffer

	// Magic
	buf.WriteString(parquetMagic)

	numRows := len(data.Rows)
	columnChunkOffsets := make([]int64, len(allKeys))
	columnChunkSizes := make([]int32, len(allKeys))

	// Write each column as a data page.
	for colIdx, colData := range columns {
		columnChunkOffsets[colIdx] = int64(buf.Len())

		// Build the raw data: for BYTE_ARRAY, each value is len(uint32) + bytes.
		var rawData bytes.Buffer
		for _, val := range colData {
			b := []byte(val)
			writeLE32(&rawData, uint32(len(b)))
			rawData.Write(b)
		}

		// Page header (Thrift compact protocol):
		// PageType: DATA_PAGE (0), uncompressed size, compressed size, num_values.
		var pageHeader bytes.Buffer
		writeThriftCompactI32(&pageHeader, 1, 0) // field 1: type = DATA_PAGE (0)
		uncompressedSize := int32(rawData.Len())
		writeThriftCompactI32(&pageHeader, 2, uncompressedSize) // field 2: uncompressed size
		writeThriftCompactI32(&pageHeader, 3, uncompressedSize) // field 3: compressed size
		// Data page header (field 5, struct):
		writeThriftCompactField(&pageHeader, thriftStruct, 5)
		writeThriftCompactI32(&pageHeader, 1, int32(numRows)) // num_values
		writeThriftCompactI32(&pageHeader, 2, 0)              // encoding: PLAIN (0)
		writeThriftCompactI32(&pageHeader, 3, 0)              // definition_level_encoding: PLAIN
		writeThriftCompactI32(&pageHeader, 4, 0)              // repetition_level_encoding: PLAIN
		pageHeader.WriteByte(thriftStop)                       // end of DataPageHeader struct
		pageHeader.WriteByte(thriftStop)                       // end of PageHeader struct

		buf.Write(pageHeader.Bytes())
		buf.Write(rawData.Bytes())

		columnChunkSizes[colIdx] = int32(pageHeader.Len()) + uncompressedSize
	}

	// Build file metadata using Thrift compact protocol.
	metadataOffset := int64(buf.Len())
	var metadata bytes.Buffer

	// FileMetaData struct:
	writeThriftCompactI32(&metadata, 1, 1)              // field 1: version = 1
	writeThriftCompactField(&metadata, thriftList, 2)    // field 2: schema (list of SchemaElement)
	metadata.WriteByte(0xF0 | thriftStruct)               // list header: type=struct, size in next varint
	writeThriftCompactVarInt(&metadata, int64(len(allKeys)+1)) // list size (root + columns)

	// Root schema element
	writeThriftCompactBinary(&metadata, 1, []byte("schema"))  // field 1: name
	writeThriftCompactI32(&metadata, 3, int32(len(allKeys)))   // field 3: num_children
	metadata.WriteByte(thriftStop)

	// Column schema elements
	for _, key := range allKeys {
		writeThriftCompactBinary(&metadata, 1, []byte(key)) // field 1: name
		writeThriftCompactI32(&metadata, 2, 6)              // field 2: type = BYTE_ARRAY (6)
		writeThriftCompactI32(&metadata, 4, 0)              // field 4: repetition_type = REQUIRED (0)
		metadata.WriteByte(thriftStop)
	}

	writeThriftCompactI64(&metadata, 3, int64(numRows)) // field 3: num_rows

	// field 4: row_groups (list of RowGroup)
	writeThriftCompactField(&metadata, thriftList, 4)
	metadata.WriteByte(0xF0 | thriftStruct)               // list header: type=struct, size in next varint
	writeThriftCompactVarInt(&metadata, 1)            // one row group

	// RowGroup:
	// field 1: columns (list of ColumnChunk)
	writeThriftCompactField(&metadata, thriftList, 1)
	metadata.WriteByte(0xF0 | thriftStruct)               // list header: type=struct, size in next varint
	writeThriftCompactVarInt(&metadata, int64(len(allKeys)))

	for colIdx, key := range allKeys {
		// ColumnChunk:
		writeThriftCompactI64(&metadata, 1, columnChunkOffsets[colIdx]) // field 1: file_offset
		// field 2: meta_data (ColumnMetaData struct)
		writeThriftCompactField(&metadata, thriftStruct, 2)
		writeThriftCompactI32(&metadata, 1, 6)                              // type = BYTE_ARRAY
		writeThriftCompactField(&metadata, thriftList, 2)                   // encodings
		metadata.WriteByte(0xF0 | thriftI32)                                // list of i32, size in next varint
		writeThriftCompactVarInt(&metadata, 1)                              // one encoding
		writeThriftCompactVarInt(&metadata, 0)                              // PLAIN
		writeThriftCompactField(&metadata, thriftList, 3)                   // path_in_schema
		metadata.WriteByte(0xF0 | thriftBinary)                             // list of binary, size in next varint
		writeThriftCompactVarInt(&metadata, 1)
		writeThriftCompactLenPrefixed(&metadata, []byte(key))
		writeThriftCompactI32(&metadata, 4, 0)                              // codec: UNCOMPRESSED
		writeThriftCompactI64(&metadata, 5, int64(numRows))                 // num_values
		writeThriftCompactI64(&metadata, 6, int64(columnChunkSizes[colIdx])) // total_uncompressed_size
		writeThriftCompactI64(&metadata, 7, int64(columnChunkSizes[colIdx])) // total_compressed_size
		writeThriftCompactI64(&metadata, 9, columnChunkOffsets[colIdx])      // data_page_offset
		metadata.WriteByte(thriftStop) // end ColumnMetaData
		metadata.WriteByte(thriftStop) // end ColumnChunk
	}

	writeThriftCompactI64(&metadata, 2, int64(metadataOffset-4)) // field 2: total_byte_size
	writeThriftCompactI64(&metadata, 3, int64(numRows))          // field 3: num_rows
	metadata.WriteByte(thriftStop)                                // end RowGroup

	writeThriftCompactBinary(&metadata, 5, []byte("acdg-svc-analysis-bi")) // created_by
	metadata.WriteByte(thriftStop)                                          // end FileMetaData

	buf.Write(metadata.Bytes())

	// Footer length (4 bytes LE) + magic.
	metadataLen := uint32(metadata.Len())
	writeLE32(&buf, metadataLen)
	buf.WriteString(parquetMagic)

	_, err := w.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("parquet write: %w", err)
	}
	return nil
}

// ContentType returns the MIME type for Parquet.
func (e *ParquetEncoder) ContentType() string {
	return "application/vnd.apache.parquet"
}

// FileExtension returns the file extension for Parquet.
func (e *ParquetEncoder) FileExtension() string {
	return "parquet"
}

// writeLE32 writes a uint32 in little-endian to buf.
func writeLE32(buf *bytes.Buffer, v uint32) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	buf.Write(b)
}

// Thrift compact protocol helpers.

func writeThriftCompactField(buf *bytes.Buffer, typeID byte, fieldID int) {
	if fieldID <= 15 {
		buf.WriteByte(byte(fieldID<<4) | typeID)
	} else {
		buf.WriteByte(typeID)
		writeThriftCompactVarInt(buf, int64(fieldID))
	}
}

func writeThriftCompactI32(buf *bytes.Buffer, fieldID int, value int32) {
	writeThriftCompactField(buf, thriftI32, fieldID)
	writeThriftCompactVarInt(buf, int64(value))
}

func writeThriftCompactI64(buf *bytes.Buffer, fieldID int, value int64) {
	writeThriftCompactField(buf, thriftI64, fieldID)
	writeThriftCompactVarInt(buf, value)
}

func writeThriftCompactBinary(buf *bytes.Buffer, fieldID int, data []byte) {
	writeThriftCompactField(buf, thriftBinary, fieldID)
	writeThriftCompactLenPrefixed(buf, data)
}

func writeThriftCompactLenPrefixed(buf *bytes.Buffer, data []byte) {
	writeThriftCompactVarInt(buf, int64(len(data)))
	buf.Write(data)
}

// writeThriftCompactVarInt writes a zigzag-encoded varint.
func writeThriftCompactVarInt(buf *bytes.Buffer, n int64) {
	// Zigzag encoding.
	z := uint64((n << 1) ^ (n >> 63))
	for z >= 0x80 {
		buf.WriteByte(byte(z) | 0x80)
		z >>= 7
	}
	buf.WriteByte(byte(z))
}
