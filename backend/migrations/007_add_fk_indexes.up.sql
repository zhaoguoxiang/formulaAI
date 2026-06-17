CREATE INDEX idx_formula_parts_formula_id ON formula_parts(formula_id);
CREATE INDEX idx_formula_parts_sort_order ON formula_parts(formula_id, sort_order);
CREATE INDEX idx_formula_ingredients_part_id ON formula_ingredients(part_id);
CREATE INDEX idx_formula_steps_formula_id ON formula_steps(formula_id);
CREATE INDEX idx_formula_steps_part_id ON formula_steps(part_id);
CREATE INDEX idx_formula_dosing_actions_step_id ON formula_dosing_actions(step_id);
CREATE INDEX idx_formula_dosing_actions_ingredient_id ON formula_dosing_actions(ingredient_id);
CREATE INDEX idx_test_items_outline_id ON test_items(outline_id);
CREATE INDEX idx_test_indicators_item_id ON test_indicators(item_id);
