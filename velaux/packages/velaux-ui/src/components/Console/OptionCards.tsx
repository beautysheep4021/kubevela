import React, { useState } from 'react';
import classNames from 'classnames';
import { AiOutlineCheck } from 'react-icons/ai';

import { Translation } from '../Translation';

export type OptionCard = {
  value: string;
  title: string;
  // group sorts the cards under tabs, as images under their frameworks.
  group?: string;
  // fields are the label-value lines the card shows.
  fields?: { label: string; value: React.ReactNode }[];
};

type Props = {
  value?: string;
  options: OptionCard[];
  onChange?: (value: string) => void;
  emptyText?: string;
};

// OptionCards picks one of many by its details: tabs by group, then a card
// for each option, the chosen card outlined and ticked in its corner.
export const OptionCards = (props: Props) => {
  const groups = Array.from(new Set(props.options.map((o) => o.group || '')));
  const chosenGroup = props.options.find((o) => o.value === props.value)?.group || '';
  const [group, setGroup] = useState<string>(groups.includes(chosenGroup) ? chosenGroup : groups[0] || '');
  const shown = props.options.filter((o) => (o.group || '') === group);
  return (
    <div className="option-cards">
      {groups.length > 1 && (
        <div className="option-cards-tabs" role="tablist">
          {groups.map((g) => (
            <span
              key={g}
              role="tab"
              aria-selected={g === group}
              className={classNames('option-cards-tab', { active: g === group })}
              onClick={() => setGroup(g)}
            >
              <Translation>{g || 'Other'}</Translation>
            </span>
          ))}
        </div>
      )}
      <div className="option-cards-list">
        {shown.length === 0 && (
          <div className="option-cards-empty">
            <Translation>{props.emptyText || 'No options'}</Translation>
          </div>
        )}
        {shown.map((o) => (
          <div
            key={o.value}
            role="radio"
            aria-checked={props.value === o.value}
            tabIndex={0}
            className={classNames('option-card', { active: props.value === o.value })}
            onClick={() => props.onChange?.(o.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                props.onChange?.(o.value);
              }
            }}
          >
            <div className="option-card-title">{o.title}</div>
            {o.fields?.map((f) => (
              <div className="option-card-field" key={f.label}>
                <span className="option-card-label">
                  <Translation>{f.label}</Translation>
                </span>
                <span className="option-card-value">{f.value}</span>
              </div>
            ))}
            {props.value === o.value && (
              <span className="option-card-check" aria-hidden="true">
                <AiOutlineCheck />
              </span>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};
