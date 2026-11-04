package nutrition

import (
	"math"
	"testing"
)

func TestQuantity(t *testing.T) {
	for _, tc := range []struct {
		name, unit  string
		input, want float64
	}{
		{"grams", "g", 12.5, 12.5}, {"kilograms", "kg", .125, 125},
		{"liters", "L", .25, 250}, {"milliliters", "ml", 30, 30},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Quantity(tc.input, tc.unit); got != tc.want {
				t.Fatalf("got %v; want %v", got, tc.want)
			}
		})
	}
}

func TestUnknownZeroAndPartialTotals(t *testing.T) {
	known := Scale(Values{"energy_kcal": Number(80), "protein_g": Number(0), "sodium_mg": Number(12)}, 250)
	unknown := Scale(Values{"energy_kcal": Number(100)}, 100)
	combined := Sum(known, unknown)
	assertAmount(t, combined["energy_kcal"], Number(300), true)
	assertAmount(t, combined["protein_g"], Number(0), false)
	assertAmount(t, combined["sodium_mg"], Number(30), false)
	assertAmount(t, combined["fat_g"], nil, false)
	assertAmount(t, combined["vitamin_d_ug"], nil, false)
	empty := Sum()
	for _, key := range Keys {
		assertAmount(t, empty[key], Number(0), true)
	}
	nested := Sum(combined, unknown)
	assertAmount(t, nested["sodium_mg"], Number(30), false)
}

func TestPrecision(t *testing.T) {
	part := Scale(Values{"energy_kcal": Number(123.4567)}, 12.3456)
	if got := *part["energy_kcal"].KnownTotal; math.Abs(got-15.2414703552) > 1e-10 {
		t.Fatalf("precision lost: %v", got)
	}
}

func TestNormalizeNutrition(t *testing.T) {
	values := Values{"energy_kcal": Number(100), "protein_g": Number(0)}
	normalized := values.Normalized()
	if len(normalized) != len(Keys) || *normalized["energy_kcal"] != 100 || *normalized["protein_g"] != 0 {
		t.Fatalf("normalization lost supplied values: %v", normalized)
	}
	for _, key := range Keys {
		if _, exists := normalized[key]; !exists {
			t.Fatalf("missing nutrient %s", key)
		}
	}
	if normalized["sodium_mg"] != nil {
		t.Fatal("unknown sodium became a known value")
	}
}

func assertAmount(t *testing.T, got Amount, want *float64, complete bool) {
	t.Helper()
	if got.Complete != complete || (got.KnownTotal == nil) != (want == nil) {
		t.Fatalf("got %+v want %v/%v", got, want, complete)
	}
	if want != nil && math.Abs(*got.KnownTotal-*want) > 1e-9 {
		t.Fatalf("got %v want %v", *got.KnownTotal, *want)
	}
}
