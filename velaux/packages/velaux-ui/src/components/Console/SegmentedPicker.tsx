import React from 'react';
import classNames from 'classnames';

import { Translation } from '../Translation';

type Option = { label: string; value: string; disabled?: boolean };

type Props = {
  // value and onChange take what a form field binds, a string in practice.
  value?: unknown;
  options: Option[];
  onChange?: (value: string) => void;
  id?: string;
};

// SegmentedPicker picks one of a few options as joined buttons, the current
// one outlined in the accent colour.
// It forwards its ref, so a form field can bind it.
export const SegmentedPicker = React.forwardRef<HTMLDivElement, Props>((props, ref) => (
  <div className="segmented-picker" role="radiogroup" ref={ref} id={props.id}>
    {props.options.map((o) => (
      <button
        key={o.value}
        type="button"
        role="radio"
        aria-checked={props.value === o.value}
        disabled={o.disabled}
        className={classNames('segmented-option', { active: props.value === o.value })}
        onClick={() => props.onChange?.(o.value)}
      >
        <Translation>{o.label}</Translation>
      </button>
    ))}
  </div>
));

SegmentedPicker.displayName = 'SegmentedPicker';
