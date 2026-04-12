package domain

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

// CSVGeographyLookup implements GeographyLookup by loading a CSV mapping
// of CEP prefixes to IBGE mesoregions. It uses progressively shorter
// prefix matching (8 down to 5 digits) to find the best match.
type CSVGeographyLookup struct {
	data map[string]Geography
}

// NewCSVGeographyLookup loads the geography mapping from a CSV file at the
// given path. The CSV must have a header row and at least 8 columns:
// cep_prefix, mesoregion_code, mesoregion_name, microregion_code,
// microregion_name, state_code, state_name, region.
func NewCSVGeographyLookup(filePath string) (*CSVGeographyLookup, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("geography: failed to open CSV: %w", err)
	}
	defer f.Close()
	return NewCSVGeographyLookupFromReader(f)
}

// NewCSVGeographyLookupFromReader loads the geography mapping from an
// io.Reader containing CSV data.
func NewCSVGeographyLookupFromReader(r io.Reader) (*CSVGeographyLookup, error) {
	reader := csv.NewReader(r)

	// Skip header row
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("geography: failed to read header: %w", err)
	}

	data := make(map[string]Geography)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("geography: failed to read row: %w", err)
		}

		// CSV columns: cep_prefix, mesoregion_code, mesoregion_name,
		//              microregion_code, microregion_name, state_code,
		//              state_name, region
		if len(record) < 8 {
			continue
		}

		cepPrefix := record[0]
		data[cepPrefix] = Geography{
			MesoregionCode:  MesoregionCode(record[1]),
			MesoregionName:  record[2],
			MicroregionCode: record[3],
			MicroregionName: record[4],
			StateCode:       record[5],
			StateName:       record[6],
			Region:          record[7],
		}
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("geography: CSV contains no data")
	}

	return &CSVGeographyLookup{data: data}, nil
}

// FindByCEP resolves a raw 8-digit CEP string to its IBGE geographic
// hierarchy. It tries progressively shorter prefixes (8, 7, 6, 5 digits)
// to find the best match.
func (g *CSVGeographyLookup) FindByCEP(cep string) (Geography, error) {
	if err := validateCEP(cep); err != nil {
		return Geography{}, err
	}

	// Try progressively shorter prefixes for best match
	for l := len(cep); l >= 5; l-- {
		if geo, ok := g.data[cep[:l]]; ok {
			return geo, nil
		}
	}

	return Geography{}, ErrCEPNotFound
}

// Verify CSVGeographyLookup satisfies GeographyLookup at compile time.
var _ GeographyLookup = (*CSVGeographyLookup)(nil)
