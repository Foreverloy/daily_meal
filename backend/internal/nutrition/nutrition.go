// Package nutrition owns nutrient definitions, unit conversion and aggregation.
// Calculations retain float64 precision; rounding belongs to the presentation layer.
package nutrition

var Keys = [...]string{"energy_kcal", "protein_g", "carbohydrate_g", "fat_g", "sodium_mg", "calcium_mg", "vitamin_c_mg", "vitamin_d_ug"}

// Values 按指标名保存营养值，单位写在 key 中，如 energy_kcal、protein_g。
// 例如 Values{"protein_g": Number(0), "sodium_mg": nil} 表示蛋白质为零、钠含量未知。
// 指针用于区分未知值 nil 和已知的 0；缺少某个 key 也按未知处理。
type Values map[string]*float64

// Amount 是一个营养指标的汇总结果，同时保留已知部分和数据完整性。
type Amount struct {
	// KnownTotal 是已知部分之和；全部未知时为 nil，无记录时的汇总为已知零。
	KnownTotal *float64 `json:"known_total"`
	// Complete 表示参与汇总的各项是否都已知；如 10g 加未知值，结果为 10 且 false。
	Complete bool `json:"complete"`
}

// Totals 按营养指标名保存汇总结果，用于食材、整道菜、餐次和全天。
// 例如 protein_g 对应的 KnownTotal 使用克，Complete 表示蛋白质数据是否完整。
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
