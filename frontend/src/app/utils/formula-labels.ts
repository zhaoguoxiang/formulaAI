import type { ComponentMode, FormulaStatus, PartName } from '../types/formula.types';

export const ModeLabelMap: Record<ComponentMode, string> = {
  single: '单组分',
  double: '双组分',
};

export const StatusLabelMap: Record<FormulaStatus, string> = {
  draft: '草稿',
  active: '已启用',
  archived: '已归档',
};

export const PartLabelMap: Record<PartName, string> = {
  PartA: 'A 组分',
  PartB: 'B 组分',
  PartMain: '主组分',
};

export function getModeLabel(mode: ComponentMode): string {
  return ModeLabelMap[mode] || mode;
}

export function getStatusLabel(status: FormulaStatus): string {
  return StatusLabelMap[status] || status;
}

export function getModeClass(mode: ComponentMode): string {
  return mode === 'single' ? 'mode-single' : 'mode-double';
}

export function getPartNameClass(name: PartName): string {
  const map: Record<PartName, string> = {
    PartA: 'part-a',
    PartB: 'part-b',
    PartMain: 'part-main',
  };
  return map[name] ?? '';
}

export function getPartNameLabel(name: PartName): string {
  return PartLabelMap[name] || name;
}
