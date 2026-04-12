package domain

import (
	"testing"
)

func TestCheckKAnonymity(t *testing.T) {
	tests := []struct {
		name           string
		groups         []IndicatorGroup
		k              int
		wantSuppressed []bool // expected BelowThreshold for each group
	}{
		{
			name:           "empty slice returns empty result",
			groups:         []IndicatorGroup{},
			k:              5,
			wantSuppressed: []bool{},
		},
		{
			name: "all groups at or above K: none suppressed",
			groups: []IndicatorGroup{
				{Labels: map[string]string{"age": "0-4"}, Value: 10, BelowThreshold: false},
				{Labels: map[string]string{"age": "5-9"}, Value: 5, BelowThreshold: false},
				{Labels: map[string]string{"age": "10-14"}, Value: 20, BelowThreshold: false},
			},
			k:              5,
			wantSuppressed: []bool{false, false, false},
		},
		{
			name: "one group below K: that group marked BelowThreshold",
			groups: []IndicatorGroup{
				{Labels: map[string]string{"age": "0-4"}, Value: 10, BelowThreshold: false},
				{Labels: map[string]string{"age": "5-9"}, Value: 3, BelowThreshold: false},
				{Labels: map[string]string{"age": "10-14"}, Value: 20, BelowThreshold: false},
			},
			k:              5,
			wantSuppressed: []bool{false, true, false},
		},
		{
			name: "all groups below K: all marked",
			groups: []IndicatorGroup{
				{Labels: map[string]string{"age": "0-4"}, Value: 1, BelowThreshold: false},
				{Labels: map[string]string{"age": "5-9"}, Value: 2, BelowThreshold: false},
				{Labels: map[string]string{"age": "10-14"}, Value: 4, BelowThreshold: false},
			},
			k:              5,
			wantSuppressed: []bool{true, true, true},
		},
		{
			name: "K=1: no groups suppressed (minimum possible K)",
			groups: []IndicatorGroup{
				{Labels: map[string]string{"age": "0-4"}, Value: 1, BelowThreshold: false},
				{Labels: map[string]string{"age": "5-9"}, Value: 100, BelowThreshold: false},
			},
			k:              1,
			wantSuppressed: []bool{false, false},
		},
		{
			name: "value exactly equal to K: not suppressed",
			groups: []IndicatorGroup{
				{Labels: map[string]string{"age": "0-4"}, Value: 5, BelowThreshold: false},
			},
			k:              5,
			wantSuppressed: []bool{false},
		},
		{
			name: "value one less than K: suppressed",
			groups: []IndicatorGroup{
				{Labels: map[string]string{"age": "0-4"}, Value: 4, BelowThreshold: false},
			},
			k:              5,
			wantSuppressed: []bool{true},
		},
		{
			name: "value zero: suppressed",
			groups: []IndicatorGroup{
				{Labels: map[string]string{"age": "0-4"}, Value: 0, BelowThreshold: false},
			},
			k:              5,
			wantSuppressed: []bool{true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckKAnonymity(tt.groups, tt.k)

			if len(result) != len(tt.wantSuppressed) {
				t.Fatalf("CheckKAnonymity returned %d groups, want %d", len(result), len(tt.wantSuppressed))
			}

			for i, want := range tt.wantSuppressed {
				if result[i].BelowThreshold != want {
					t.Errorf("group[%d].BelowThreshold = %v, want %v (Value=%d, K=%d)",
						i, result[i].BelowThreshold, want, result[i].Value, tt.k)
				}
			}
		})
	}
}

func TestCheckKAnonymity_DoesNotMutateInput(t *testing.T) {
	original := []IndicatorGroup{
		{Labels: map[string]string{"age": "0-4"}, Value: 2, BelowThreshold: false},
		{Labels: map[string]string{"age": "5-9"}, Value: 10, BelowThreshold: false},
	}

	// Save original BelowThreshold values
	origBT0 := original[0].BelowThreshold
	origBT1 := original[1].BelowThreshold

	result := CheckKAnonymity(original, 5)

	// The result should have the first group suppressed
	if !result[0].BelowThreshold {
		t.Errorf("result[0].BelowThreshold = false, want true (Value=2, K=5)")
	}

	// But the original slice must NOT be mutated
	if original[0].BelowThreshold != origBT0 {
		t.Errorf("original[0].BelowThreshold was mutated from %v to %v", origBT0, original[0].BelowThreshold)
	}
	if original[1].BelowThreshold != origBT1 {
		t.Errorf("original[1].BelowThreshold was mutated from %v to %v", origBT1, original[1].BelowThreshold)
	}
}

func TestCheckKAnonymity_PreservesLabelsAndValues(t *testing.T) {
	groups := []IndicatorGroup{
		{Labels: map[string]string{"age": "0-4", "sex": "MALE"}, Value: 3, BelowThreshold: false},
		{Labels: map[string]string{"age": "5-9", "sex": "FEMALE"}, Value: 10, BelowThreshold: false},
	}

	result := CheckKAnonymity(groups, 5)

	if result[0].Value != 3 {
		t.Errorf("result[0].Value = %d, want 3", result[0].Value)
	}
	if result[0].Labels["age"] != "0-4" {
		t.Errorf("result[0].Labels[age] = %q, want %q", result[0].Labels["age"], "0-4")
	}
	if result[0].Labels["sex"] != "MALE" {
		t.Errorf("result[0].Labels[sex] = %q, want %q", result[0].Labels["sex"], "MALE")
	}
	if result[1].Value != 10 {
		t.Errorf("result[1].Value = %d, want 10", result[1].Value)
	}
}

func TestCheckKAnonymity_UsesKThresholdConstant(t *testing.T) {
	// Verify the KThreshold constant is 5 as documented in the contract
	if KThreshold != 5 {
		t.Errorf("KThreshold = %d, want 5", KThreshold)
	}
}

func TestCountSuppressed(t *testing.T) {
	tests := []struct {
		name   string
		groups []IndicatorGroup
		want   int
	}{
		{
			name:   "empty slice returns 0",
			groups: []IndicatorGroup{},
			want:   0,
		},
		{
			name: "no suppressed groups returns 0",
			groups: []IndicatorGroup{
				{Value: 10, BelowThreshold: false},
				{Value: 20, BelowThreshold: false},
				{Value: 30, BelowThreshold: false},
			},
			want: 0,
		},
		{
			name: "3 out of 5 suppressed returns 3",
			groups: []IndicatorGroup{
				{Value: 1, BelowThreshold: true},
				{Value: 10, BelowThreshold: false},
				{Value: 2, BelowThreshold: true},
				{Value: 20, BelowThreshold: false},
				{Value: 3, BelowThreshold: true},
			},
			want: 3,
		},
		{
			name: "all suppressed returns count equal to len",
			groups: []IndicatorGroup{
				{Value: 1, BelowThreshold: true},
				{Value: 2, BelowThreshold: true},
				{Value: 3, BelowThreshold: true},
			},
			want: 3,
		},
		{
			name: "single suppressed group returns 1",
			groups: []IndicatorGroup{
				{Value: 1, BelowThreshold: true},
			},
			want: 1,
		},
		{
			name: "single non-suppressed group returns 0",
			groups: []IndicatorGroup{
				{Value: 10, BelowThreshold: false},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CountSuppressed(tt.groups)
			if got != tt.want {
				t.Errorf("CountSuppressed() = %d, want %d", got, tt.want)
			}
		})
	}
}
