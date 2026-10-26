import type { Tone } from '../../components/StatusBadge';

export const statusList = [
  { label: 'Init', value: 'init' },
  { label: 'Running', value: 'running' },
  { label: 'Complete', value: 'complete' },
  { label: 'Failure', value: 'failure' },
  { label: 'Terminated', value: 'terminated' },
  { label: 'Rollback', value: 'rollback' },
];

// statusTones colours each revision status as the badges do.
export const statusTones: Record<string, Tone> = {
  init: 'neutral',
  running: 'progressing',
  complete: 'healthy',
  failure: 'failed',
  terminated: 'suspended',
  rollback: 'neutral',
};
