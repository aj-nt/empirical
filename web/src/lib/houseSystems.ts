// House systems supported by the backend (sweph codes)
export const HOUSE_SYSTEMS: { value: string; label: string }[] = [
  { value: 'placidus', label: 'Placidus' },
  { value: 'whole_sign', label: 'Whole Sign' },
  { value: 'equal', label: 'Equal' },
  { value: 'porphyry', label: 'Porphyry' },
  { value: 'koch', label: 'Koch' },
  { value: 'regiomontanus', label: 'Regiomontanus' },
  { value: 'alcabitius', label: 'Alcabitius' },
  { value: 'campanus', label: 'Campanus' },
];

export const DEFAULT_HOUSE_SYSTEM = 'placidus';
