package domain

import (
	"fmt"
	"testing"
	"time"
)

// --- HashPatientID benchmarks ---

func BenchmarkHashPatientID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = HashPatientID("patient-abc-123", "production-salt-value")
	}
}

func BenchmarkHashPatientID_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = HashPatientID("patient-abc-123", "production-salt-value")
		}
	})
}

// --- GeneralizeAge benchmarks ---

func BenchmarkGeneralizeAge(b *testing.B) {
	birth := time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC)
	ref := time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GeneralizeAge(birth, ref)
	}
}

func BenchmarkGeneralizeAge_AllBands(b *testing.B) {
	ref := time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC)
	births := []time.Time{
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),  // 0-4
		time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),  // 5-9
		time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),  // 25-29
		time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),  // 55-59
		time.Date(1940, 1, 1, 0, 0, 0, 0, time.UTC),  // 80+
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GeneralizeAge(births[i%len(births)], ref)
	}
}

// --- GeneralizeIncome benchmarks ---

func BenchmarkGeneralizeIncome(b *testing.B) {
	incomes := []int64{0, 35000, 100000, 200000, 400000, 800000}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GeneralizeIncome(incomes[i%len(incomes)])
	}
}

// --- CheckKAnonymity benchmarks ---

func BenchmarkCheckKAnonymity_10Groups(b *testing.B) {
	groups := makeGroups(10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CheckKAnonymity(groups, KThreshold)
	}
}

func BenchmarkCheckKAnonymity_100Groups(b *testing.B) {
	groups := makeGroups(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CheckKAnonymity(groups, KThreshold)
	}
}

func BenchmarkCheckKAnonymity_1000Groups(b *testing.B) {
	groups := makeGroups(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CheckKAnonymity(groups, KThreshold)
	}
}

func BenchmarkCheckKAnonymity_10000Groups(b *testing.B) {
	groups := makeGroups(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CheckKAnonymity(groups, KThreshold)
	}
}

func BenchmarkCountSuppressed_1000Groups(b *testing.B) {
	groups := CheckKAnonymity(makeGroups(1000), KThreshold)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CountSuppressed(groups)
	}
}

// --- Period benchmarks ---

func BenchmarkNewPeriod(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewPeriod(2026, (i%12)+1)
	}
}

func BenchmarkPeriod_YearMonth(b *testing.B) {
	p := Period{Year: 2026, Month: 4}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.YearMonth()
	}
}

func BenchmarkPeriod_Quarter(b *testing.B) {
	p := Period{Year: 2026, Month: 7}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Quarter()
	}
}

// --- Allocation benchmarks ---

func BenchmarkHashPatientID_Allocs(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = HashPatientID("patient-abc-123", "production-salt-value")
	}
}

func BenchmarkCheckKAnonymity_100Groups_Allocs(b *testing.B) {
	groups := makeGroups(100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CheckKAnonymity(groups, KThreshold)
	}
}

func BenchmarkGeneralizeAge_Allocs(b *testing.B) {
	birth := time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC)
	ref := time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GeneralizeAge(birth, ref)
	}
}

// --- helpers ---

func makeGroups(n int) []IndicatorGroup {
	groups := make([]IndicatorGroup, n)
	for i := range groups {
		groups[i] = IndicatorGroup{
			Labels: map[string]string{
				"age_band":   fmt.Sprintf("band-%d", i%17),
				"sex":        "MALE",
				"mesoregion": fmt.Sprintf("%04d", i%100),
			},
			Value:          (i % 10) + 1, // values 1-10, some below K=5
			BelowThreshold: false,
		}
	}
	return groups
}
