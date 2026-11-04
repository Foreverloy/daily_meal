// Package nutrition owns nutrient definitions, unit conversion and aggregation.
// Calculations retain float64 precision; rounding belongs to the presentation layer.
package nutrition

var Keys = [...]string{"energy_kcal", "protein_g", "carbohydrate_g", "fat_g", "sodium_mg", "calcium_mg", "vitamin_c_mg", "vitamin_d_ug"}

type Values map[string]*float64

type Amount struct {
	KnownTotal *float64 `json:"known_total"`
	Complete   bool     `json:"complete"`
}

type Totals map[string]Amount

func Number(v float64) *float64 { return &v }

// Normalized returns every supported nutrient, including explicit unknowns.
func (v Values) Normalized() Values {
	out := make(Values, len(Keys))
	for _, k := range Keys {
		out[k] = v[k]
	}
	return out
}

// Quantity 将 kg/L 换算为 g/ml；调用方提供与食物基准相符的单位。
func Quantity(quantity float64, unit string) float64 {
	// TODO：暂未校验数量范围与单位是否同类。
	if unit == "kg" || unit == "L" {
		quantity *= 1000
	}
	return quantity
}

// TODO：Scale 和 Sum 暂未检查计算溢出，调用方须提供正常范围的营养值与用量。
func Scale(values Values, quantity float64) Totals {
	out := make(Totals, len(Keys))
	for _, k := range Keys {
		a := Amount{}
		if n := values[k]; n != nil {
			v := *n * (quantity / 100)
			a = Amount{KnownTotal: Number(v), Complete: true}
		}
		out[k] = a
	}
	return out
}

// Sum preserves completeness through nested aggregation (ingredients -> meals -> day).
func Sum(parts ...Totals) Totals {
	out := make(Totals, len(Keys))
	for _, k := range Keys {
		total, known, complete := 0.0, len(parts) == 0, true
		for _, part := range parts {
			a := part[k]
			complete = complete && a.Complete
			if a.KnownTotal != nil {
				total += *a.KnownTotal
				known = true
			}
		}
		a := Amount{Complete: complete}
		if known {
			a.KnownTotal = Number(total)
		}
		out[k] = a
	}
	return out
}
