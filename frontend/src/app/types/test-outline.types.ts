/** Single indicator within a test item (mirrors Go TestIndicator). */
export interface TestIndicator {
  id: string;
  item_id: string;
  name: string;
  unit?: string;
  min_value?: number;
  max_value?: number;
  sample_prep_method?: string;
  test_method?: string;
  test_condition?: string;
  version?: string;
  sort_order: number;
  created_at: string;
}

/** A detection item grouping related indicators (mirrors Go TestItem). */
export interface TestItem {
  id: string;
  outline_id: string;
  name: string;
  sort_order: number;
  indicators: TestIndicator[];
  created_at: string;
}

/** Top-level test outline with versioning (mirrors Go TestOutline). */
export interface TestOutline {
  id: string;
  name: string;
  version: number;
  status: OutlineStatus;
  items: TestItem[];
  created_at: string;
  updated_at: string;
}

/** Lifecycle status of a test outline. */
export type OutlineStatus = 'draft' | 'active' | 'archived';
