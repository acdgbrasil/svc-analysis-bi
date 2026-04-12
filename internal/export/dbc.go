package export

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"fmt"
	"io"
)

// DBCEncoder encodes ExportData as DBC format: DBF compressed with PKZIP
// deflate (compatible with DataSUS tooling).
//
// DBC file structure:
//   - 8-byte header: original size (4 bytes LE) + compressed size placeholder (4 bytes LE)
//   - deflate-compressed DBF data
//
// DataSUS DBC files historically use blast/LZ77, but modern tooling accepts
// standard deflate. We use Go's compress/flate for portability.
type DBCEncoder struct{}

// dbcHeaderSize is the size of the DBC file header.
const dbcHeaderSize = 8

// Encode writes the export data as a DBC (compressed DBF) file to w.
func (e *DBCEncoder) Encode(w io.Writer, data ExportData) error {
	// First encode as DBF.
	dbfData, err := EncodeDBFToBytes(data)
	if err != nil {
		return fmt.Errorf("dbc encode dbf: %w", err)
	}

	// Compress with deflate.
	var compressed bytes.Buffer
	flateWriter, err := flate.NewWriter(&compressed, flate.BestCompression)
	if err != nil {
		return fmt.Errorf("dbc flate init: %w", err)
	}
	if _, err := flateWriter.Write(dbfData); err != nil {
		return fmt.Errorf("dbc flate write: %w", err)
	}
	if err := flateWriter.Close(); err != nil {
		return fmt.Errorf("dbc flate close: %w", err)
	}

	// Build the DBC output: header + compressed data.
	var buf bytes.Buffer

	// Header: original size (4 bytes LE).
	origSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(origSize, uint32(len(dbfData)))
	buf.Write(origSize)

	// Header: compressed size (4 bytes LE).
	compSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(compSize, uint32(compressed.Len()))
	buf.Write(compSize)

	// Compressed data.
	buf.Write(compressed.Bytes())

	if _, err := w.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("dbc write: %w", err)
	}
	return nil
}

// ContentType returns the MIME type for DBC.
func (e *DBCEncoder) ContentType() string {
	return "application/x-dbc"
}

// FileExtension returns the file extension for DBC.
func (e *DBCEncoder) FileExtension() string {
	return "dbc"
}
