package nutrition

import (
	"math"
	"testing"
)

// testKeys 模拟数据库目录，包含原有指标及新增的膳食纤维、钾、硒。
var testKeys = []string{"energy_kcal", "protein_g", "carbohydrate_g", "fat_g", "sodium_mg", "calcium_mg", "vitamin_c_mg", "vitamin_d_ug", "dietary_fiber_g", "potassium_mg", "selenium_ug"}

// TestQuantity 验证 g/kg、ml/L 输入用量到 g/ml 的换算。
func TestQuantity(t *testing.T) {
	for _, tc := range []struct {
		// name 是测试名称，unit 是输入用量的单位。
		name, unit string
		// input 是输入用量，want 是预期换算结果，如 0.125kg 对应 125g。
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

// TestUnknownZeroAndPartialTotals 验证新旧指标在未知、零值和嵌套汇总中的完整性。
func TestUnknownZeroAndPartialTotals(t *testing.T) {
	known := Scale(Values{"energy_kcal": Number(80), "protein_g": Number(0), "sodium_mg": Number(12)}, 250)
	unknown := Scale(Values{"energy_kcal": Number(100)}, 100)
	combined := Sum(testKeys, known, unknown)
	assertAmount(t, combined["energy_kcal"], Number(300), true)
	assertAmount(t, combined["protein_g"], Number(0), false)
	assertAmount(t, combined["sodium_mg"], Number(30), false)
	assertAmount(t, combined["fat_g"], nil, false)
	assertAmount(t, combined["vitamin_d_ug"], nil, false)
	empty := Sum(testKeys)
	for _, key := range testKeys {
		assertAmount(t, empty[key], Number(0), true)
	}
	nested := Sum(testKeys, combined, unknown)
	assertAmount(t, nested["sodium_mg"], Number(30), false)
}

// TestPrecision 验证缩放计算保留中间精度。
func TestPrecision(t *testing.T) {
	part := Scale(Values{"energy_kcal": Number(123.4567)}, 12.3456)
	if got := *part["energy_kcal"].KnownTotal; math.Abs(got-15.2414703552) > 1e-10 {
		t.Fatalf("precision lost: %v", got)
	}
}

// TestNormalizeNutrition 验证按外部目录补齐指标且不把未知补成零。
func TestNormalizeNutrition(t *testing.T) {
	values := Values{"energy_kcal": Number(100), "protein_g": Number(0)}
	normalized := values.Normalized(testKeys)
	if len(normalized) != len(testKeys) || *normalized["energy_kcal"] != 100 || *normalized["protein_g"] != 0 {
		t.Fatalf("normalization lost supplied values: %v", normalized)
	}
	for _, key := range testKeys {
		if _, exists := normalized[key]; !exists {
			t.Fatalf("missing nutrient %s", key)
		}
	}
	if normalized["sodium_mg"] != nil {
		t.Fatal("unknown sodium became a known value")
	}
}

// assertAmount 核对营养合计的数值和完整性。
func assertAmount(t *testing.T, got Amount, want *float64, complete bool) {
	t.Helper()
	if got.Complete != complete || (got.KnownTotal == nil) != (want == nil) {
		t.Fatalf("got %+v want %v/%v", got, want, complete)
	}
	if want != nil && math.Abs(*got.KnownTotal-*want) > 1e-9 {
		t.Fatalf("got %v want %v", *got.KnownTotal, *want)
	}
}

// TestExtendedNutrients 验证新增目录指标参与缩放，旧快照缺项仍使汇总不完整。
func TestExtendedNutrients(t *testing.T) {
	current := Scale(Values{"dietary_fiber_g": Number(4), "potassium_mg": Number(200), "selenium_ug": Number(0)}, 150)
	old := Scale(Values{"energy_kcal": Number(100)}, 100)
	total := Sum(testKeys, current, old)
	assertAmount(t, total["dietary_fiber_g"], Number(6), false)
	assertAmount(t, total["potassium_mg"], Number(300), false)
	assertAmount(t, total["selenium_ug"], Number(0), false)
}
