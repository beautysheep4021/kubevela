import React from 'react';
import { Input } from '@alifd/next';
import { AiOutlineMinusCircle, AiOutlinePlus } from 'react-icons/ai';

import i18n from '../../i18n';
import { Translation } from '../Translation';

export type KV = { key: string; value: string };

type Props = {
  value?: KV[];
  onChange?: (value: KV[]) => void;
  keyPlaceholder?: string;
  valuePlaceholder?: string;
  addLabel?: string;
};

// KVList edits pairs as rows of key : value, added with the plus and removed
// with the minus at each row's end.
export const KVList = (props: Props) => {
  const rows = props.value || [];
  const set = (next: KV[]) => props.onChange?.(next);
  return (
    <div className="kv-list">
      {rows.map((row, i) => (
        <div className="kv-row" key={i}>
          <Input
            value={row.key}
            placeholder={i18n.t(props.keyPlaceholder || 'Key').toString()}
            onChange={(v: string) => set(rows.map((r, j) => (j === i ? { ...r, key: v } : r)))}
          />
          <span className="kv-sep">:</span>
          <Input
            value={row.value}
            placeholder={i18n.t(props.valuePlaceholder || 'Value').toString()}
            onChange={(v: string) => set(rows.map((r, j) => (j === i ? { ...r, value: v } : r)))}
          />
          <span
            className="kv-remove"
            role="button"
            aria-label={i18n.t('Remove').toString()}
            onClick={() => set(rows.filter((_, j) => j !== i))}
          >
            <AiOutlineMinusCircle />
          </span>
        </div>
      ))}
      <span className="kv-add" role="button" onClick={() => set([...rows, { key: '', value: '' }])}>
        <AiOutlinePlus />
        <Translation>{props.addLabel || 'Add'}</Translation>
      </span>
    </div>
  );
};

// toRecord turns the pairs into a map, the blank keys dropped.
export function toRecord(rows: KV[]): Record<string, string> {
  const out: Record<string, string> = {};
  rows.forEach((r) => {
    if (r.key.trim()) {
      out[r.key.trim()] = r.value;
    }
  });
  return out;
}
