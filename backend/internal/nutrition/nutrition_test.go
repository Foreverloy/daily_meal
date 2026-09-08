package nutrition

import (
	"math"
	"testing"
)

func TestQuantity(t *testing.T) {
	for _, tc := range []struct {
		name, unit, basis string
		input, want       float64
		invalid           bool
	}{
		{"grams", "g", "g", 12.5, 12.5, false}, {"kilograms", "kg", "g", .125, 125, false},
		{"liters", "L", "ml", .25, 250, false}, {"milliliters", "ml", "ml", 30, 30, false},
		{"cross dimension", "g", "ml", 100, 0, true}, {"serving", "份", "g", 1, 0, true},
		{"lowercase liter", "l", "ml", 1, 0, true}, {"zero", "g", "g", 0, 0, true},
		{"negative", "g", "g", -1, 0, true}, {"overflow", "kg", "g", math.MaxFloat64, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Quantity(tc.input, tc.unit, tc.basis)
			if (err != nil) != tc.invalid || got != tc.want {
				t.Fatalf("got %v, %v; want %v, invalid=%v", got, err, tc.want, tc.invalid)
			}
		})
	}
}

func TestUnknownZeroAndPartialTotals(t *testing.T) {
	known, err := Scale(Values{"energy_kcal": Number(80), "protein_g": Number(0), "sodium_mg": Number(12)}, 250)
	if err != nil {
		t.Fatal(err)
	}
	unknown, err := Scale(Values{"energy_kcal": Number(100)}, 100)
	if err != nil {
		t.Fatal(err)
	}
	combined, err := Sum(known, unknown)
	if err != nil {
		t.Fatal(err)
	}
	assertAmount(t, combined["energy_kcal"], Number(300), true)
	assertAmount(t, combined["protein_g"], Number(0), false)
	assertAmount(t, combined["sodium_mg"], Number(30), false)
	assertAmount(t, combined["fat_g"], nil, false)
	assertAmount(t, combined["vitamin_d_ug"], nil, false)
	empty, err := Sum()
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range Keys {
		assertAmount(t, empty[key], Number(0), true)
	}
	nested, err := Sum(combined, unknown)
	if err != nil {
		t.Fatal(err)
	}
	assertAmount(t, nested["sodium_mg"], Number(30), false)
}

func TestPrecisionAndOverflow(t *testing.T) {
	part, err := Scale(Values{"energy_kcal": Number(123.4567)}, 12.3456)
	if err != nil {
		t.Fatal(err)
	}
	if got := *part["energy_kcal"].KnownTotal; math.Abs(got-15.2414703552) > 1e-10 {
		t.Fatalf("precision lost: %v", got)
	}
	if _, err := Scale(Values{"energy_kcal": Number(math.MaxFloat64)}, 1000); err == nil {
		t.Fatal("expected overflow")
	}
	large := Totals{"energy_kcal": {KnownTotal: Number(math.MaxFloat64), Complete: true}}
	if _, err := Sum(large, large); err == nil {
		t.Fatal("expected sum overflow")
	}
}

func TestValidateNutrition(t *testing.T) {
	for _, values := range []Values{nil, {"energy_kcal": nil}, {"energy_kcal": Number(-1)}, {"energy_kcal": Number(0), "unknown": nil}, {"energy_kcal": Number(0), "fat_g": Number(math.Inf(1))}} {
		if err := values.Validate(); err == nil {
			t.Fatalf("accepted invalid nutrition: %v", values)
		}
	}
	if err := (Values{"energy_kcal": Number(0), "fat_g": nil}).Validate(); err != nil {
		t.Fatal(err)
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
