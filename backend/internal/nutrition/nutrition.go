// Package nutrition owns nutrient definitions, unit conversion and aggregation.
// Calculations retain float64 precision; rounding belongs to the presentation layer.
package nutrition

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

// Number 将已知数值转为指针，以便与未知值 nil 区分。
func Number(v float64) *float64 { return &v }

// Normalized 按传入的营养素目录补齐未知值，保留已知的零。
func (v Values) Normalized(keys []string) Values {
	out := make(Values, len(keys))
	for _, k := range keys {
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

// Scale 按食物实际用量缩放传入的营养数据，也用于原有历史快照。
// TODO：Scale 和 Sum 暂未检查计算溢出，调用方须提供正常范围的营养值与用量。
func Scale(values Values, quantity float64) Totals {
	out := make(Totals, len(values))
	for k, n := range values {
		a := Amount{}
		if n != nil {
			v := *n * (quantity / 100)
			a = Amount{KnownTotal: Number(v), Complete: true}
		}
		out[k] = a
	}
	return out
}

// Sum 按目录汇总食材、餐次或全天营养；旧快照缺少新指标时保持未知及不完整。
func Sum(keys []string, parts ...Totals) Totals {
	out := make(Totals, len(keys))
	for _, k := range keys {
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
