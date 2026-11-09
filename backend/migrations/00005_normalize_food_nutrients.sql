-- +goose Up
-- 营养素代码与标准单位分别保存；接口键仍为 code_unit，历史快照因此无需改名。
-- 常见指标及计量口径参考 USDA / NHANES：
-- https://wwwn.cdc.gov/Nchs/Data/Nhanes/Public/2021/DataFiles/DRXFCD_L.htm
-- https://www.fda.gov/food/nutrition-facts-label/daily-value-nutrition-and-supplement-facts-labels
-- RAE、α-生育酚和总叶酸采用明确口径，不将 IU、RE、DFE 当成普通质量单位换算。
CREATE TABLE nutrients (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    unit TEXT NOT NULL
);

INSERT INTO nutrients (code, name, unit) VALUES
    ('energy', '能量', 'kcal'),
    ('protein', '蛋白质', 'g'),
    ('carbohydrate', '碳水化合物', 'g'),
    ('fat', '脂肪', 'g'),
    ('dietary_fiber', '膳食纤维', 'g'),
    ('sugar', '总糖', 'g'),
    ('added_sugar', '添加糖', 'g'),
    ('saturated_fat', '饱和脂肪酸', 'g'),
    ('monounsaturated_fat', '单不饱和脂肪酸', 'g'),
    ('polyunsaturated_fat', '多不饱和脂肪酸', 'g'),
    ('trans_fat', '反式脂肪酸', 'g'),
    ('cholesterol', '胆固醇', 'mg'),
    ('water', '水分', 'g'),
    ('sodium', '钠', 'mg'),
    ('potassium', '钾', 'mg'),
    ('calcium', '钙', 'mg'),
    ('phosphorus', '磷', 'mg'),
    ('magnesium', '镁', 'mg'),
    ('iron', '铁', 'mg'),
    ('zinc', '锌', 'mg'),
    ('copper', '铜', 'mg'),
    ('manganese', '锰', 'mg'),
    ('selenium', '硒', 'ug'),
    ('iodine', '碘', 'ug'),
    ('vitamin_a_rae', '维生素 A（RAE）', 'ug'),
    ('vitamin_b1', '维生素 B1（硫胺素）', 'mg'),
    ('vitamin_b2', '维生素 B2（核黄素）', 'mg'),
    ('niacin', '烟酸', 'mg'),
    ('pantothenic_acid', '泛酸', 'mg'),
    ('vitamin_b6', '维生素 B6', 'mg'),
    ('biotin', '生物素', 'ug'),
    ('folate_total', '总叶酸', 'ug'),
    ('vitamin_b12', '维生素 B12', 'ug'),
    ('vitamin_c', '维生素 C', 'mg'),
    ('vitamin_d', '维生素 D', 'ug'),
    ('vitamin_e_alpha', '维生素 E（α-生育酚）', 'mg'),
    ('vitamin_k', '维生素 K', 'ug'),
    ('choline', '胆碱', 'mg');

CREATE TABLE food_nutrients (
    food_id BIGINT NOT NULL REFERENCES foods(id) ON DELETE CASCADE,
    nutrient_id BIGINT NOT NULL REFERENCES nutrients(id),
    amount_per_100 NUMERIC,
    PRIMARY KEY (food_id, nutrient_id)
);

-- 迁移所有旧食物；JSON null 和未提供的指标均保留为 SQL NULL，0 仍为 0。
INSERT INTO food_nutrients (food_id, nutrient_id, amount_per_100)
SELECT f.id, n.id, (f.nutrition_per_100 ->> (n.code || '_' || n.unit))::numeric
FROM foods f CROSS JOIN nutrients n;

ALTER TABLE foods DROP COLUMN nutrition_per_100;

-- +goose Down
-- 回退时将当前全部关联含量重新组成 JSONB，不丢弃扩展营养素或修改历史快照。
ALTER TABLE foods ADD COLUMN nutrition_per_100 JSONB;
UPDATE foods f SET nutrition_per_100 = (
    SELECT jsonb_object_agg(n.code || '_' || n.unit, fn.amount_per_100)
    FROM food_nutrients fn JOIN nutrients n ON n.id = fn.nutrient_id
    WHERE fn.food_id = f.id
);
ALTER TABLE foods ALTER COLUMN nutrition_per_100 SET NOT NULL;
ALTER TABLE foods ADD CONSTRAINT foods_nutrition_per_100_check CHECK (
    jsonb_typeof(nutrition_per_100) = 'object' AND
    nutrition_per_100 ? 'energy_kcal' AND
    jsonb_typeof(nutrition_per_100->'energy_kcal') = 'number'
);
DROP TABLE food_nutrients;
DROP TABLE nutrients;
