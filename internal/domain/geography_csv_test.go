package domain

import (
	"errors"
	"strings"
	"testing"
)

const testCSV = `cep_prefix,mesoregion_code,mesoregion_name,microregion_code,microregion_name,state_code,state_name,region
01000,3550,Metropolitana de Sao Paulo,35051,Sao Paulo,35,Sao Paulo,Sudeste
20000,3301,Metropolitana do Rio de Janeiro,33001,Rio de Janeiro,33,Rio de Janeiro,Sudeste
70000,5300,Distrito Federal,53001,Brasilia,53,Distrito Federal,Centro-Oeste
`

func TestCSVGeographyLookup_LoadFromReader(t *testing.T) {
	lookup, err := NewCSVGeographyLookupFromReader(strings.NewReader(testCSV))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lookup == nil {
		t.Fatal("expected non-nil lookup")
	}
	if len(lookup.data) != 3 {
		t.Errorf("expected 3 entries, got %d", len(lookup.data))
	}
}

func TestCSVGeographyLookup_FindByCEP_KnownPrefix(t *testing.T) {
	lookup, err := NewCSVGeographyLookupFromReader(strings.NewReader(testCSV))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// "01000000" starts with "01000" which is in the CSV
	geo, err := lookup.FindByCEP("01000000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if geo.MesoregionCode != "3550" {
		t.Errorf("expected mesoregion code 3550, got %s", geo.MesoregionCode)
	}
	if geo.MesoregionName != "Metropolitana de Sao Paulo" {
		t.Errorf("expected Metropolitana de Sao Paulo, got %s", geo.MesoregionName)
	}
	if geo.StateCode != "35" {
		t.Errorf("expected state code 35, got %s", geo.StateCode)
	}
	if geo.Region != "Sudeste" {
		t.Errorf("expected region Sudeste, got %s", geo.Region)
	}
}

func TestCSVGeographyLookup_FindByCEP_UnknownPrefix(t *testing.T) {
	lookup, err := NewCSVGeographyLookupFromReader(strings.NewReader(testCSV))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = lookup.FindByCEP("99999999")
	if !errors.Is(err, ErrCEPNotFound) {
		t.Errorf("expected ErrCEPNotFound, got: %v", err)
	}
}

func TestCSVGeographyLookup_FindByCEP_InvalidFormat(t *testing.T) {
	lookup, err := NewCSVGeographyLookupFromReader(strings.NewReader(testCSV))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name string
		cep  string
		want error
	}{
		{"too short", "1234", ErrCEPWrongLength},
		{"too long", "123456789", ErrCEPWrongLength},
		{"non-digit", "0100A000", ErrCEPNonDigit},
		{"empty", "", ErrCEPWrongLength},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := lookup.FindByCEP(tt.cep)
			if !errors.Is(err, tt.want) {
				t.Errorf("FindByCEP(%q) = %v, want %v", tt.cep, err, tt.want)
			}
		})
	}
}

func TestCSVGeographyLookup_EmptyCSV(t *testing.T) {
	emptyCSV := "cep_prefix,mesoregion_code,mesoregion_name,microregion_code,microregion_name,state_code,state_name,region\n"
	_, err := NewCSVGeographyLookupFromReader(strings.NewReader(emptyCSV))
	if err == nil {
		t.Fatal("expected error for empty CSV data")
	}
}

func TestCSVGeographyLookup_NoHeader(t *testing.T) {
	_, err := NewCSVGeographyLookupFromReader(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for CSV with no header")
	}
}
