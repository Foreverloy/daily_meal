// Package nutrition owns nutrient definitions, unit conversion and aggregation.
// Calculations retain float64 precision; rounding belongs to the presentation layer.
package nutrition

import (
	"fmt"
	"math"
)

var Keys = [...]string{"energy_kcal", "protein_g", "carbohydrate_g", "fat_g", "sodium_mg", "calcium_mg", "vitamin_c_mg", "vitamin_d_ug"}

type Values map[string]*float64

type Amount struct {
	KnownTotal *float64 `json:"known_total"`
	Complete   bool     `json:"complete"`
}

type Totals map[string]Amount

func Number(v float64) *float64 { return &v }

func Finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func Supported(key string) bool {
	for _, k := range Keys {
		if k == key {
			return true
		}
	}
	return false
}

func (v Values) Validate() error {
	if v["energy_kcal"] == nil {
		return fmt.Errorf("energy_kcal 必填")
	}
	for k, n := range v {
		if !Supported(k) {
			return fmt.Errorf("不支持营养字段 %s", k)
		}
		if n != nil && (!Finite(*n) || *n < 0) {
			return fmt.Errorf("%s 必须是非负有限数值", k)
		}
	}
	return nil
}

// Normalized returns every supported nutrient, including explicit unknowns.
func (v Values) Normalized() Values {
	out := make(Values, len(Keys))
	for _, k := range Keys {
		out[k] = v[k]
	}
	return out
}

func Quantity(quantity float64, unit, basis string) (float64, error) {
	if !Finite(quantity) || quantity <= 0 {
		return 0, fmt.Errorf("数量必须大于 0 且为有限数值")
	}
	if (basis == "g" && unit == "kg") || (basis == "ml" && unit == "L") {
		quantity *= 1000
	} else if (basis != "g" && basis != "ml") || unit != basis {
		return 0, fmt.Errorf("单位 %s 不适用于基准 %s", unit, basis)
	}
	if !Finite(quantity) {
		return 0, fmt.Errorf("数量超出计算范围")
	}
	return quantity, nil
}

func Scale(values Values, quantity float64) (Totals, error) {
	out := make(Totals, len(Keys))
	for _, k := range Keys {
		a := Amount{}
		if n := values[k]; n != nil {
			v := *n * (quantity / 100)
			if !Finite(v) {
				return nil, fmt.Errorf("营养计算超出范围")
			}
			a = Amount{KnownTotal: Number(v), Complete: true}
		}
		out[k] = a
	}
	return out, nil
}

// Sum preserves completeness through nested aggregation (ingredients -> meals -> day).
func Sum(parts ...Totals) (Totals, error) {
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
		if !Finite(total) {
			return nil, fmt.Errorf("营养合计超出范围")
		}
		a := Amount{Complete: complete}
		if known {
			a.KnownTotal = Number(total)
		}
		out[k] = a
	}
	return out, nil
}
