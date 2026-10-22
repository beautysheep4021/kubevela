import React from 'react';
import { Dropdown, Menu } from '@alifd/next';
import { AiOutlineDown } from 'react-icons/ai';
import classNames from 'classnames';

import i18n from '../../i18n';
import { splitLinks } from './links';

export type RowLink = {
  label: string;
  onClick?: () => void;
  disabled?: boolean;
  // reason says why a disabled link cannot be used.
  reason?: string;
  danger?: boolean;
  hidden?: boolean;
  // more folds the link under More whatever room there is.
  more?: boolean;
};

type Props = {
  links: RowLink[];
  // visible is how many links show before the rest fold into More.
  visible?: number;
};

// RowLinks is a row's operations as text links, the first few shown and the
// rest under More, as the cloud consoles' tables give them.
export const RowLinks = (props: Props) => {
  const { shown, folded } = splitLinks(props.links, props.visible ?? 3);
  const label = (l: RowLink) => i18n.t(l.label).toString();
  return (
    <span className="row-links">
      {shown.map((l) => (
        <button
          key={l.label}
          type="button"
          className={classNames('row-link', { danger: l.danger })}
          disabled={l.disabled}
          title={l.disabled && l.reason ? i18n.t(l.reason).toString() : undefined}
          onClick={l.onClick}
        >
          {label(l)}
        </button>
      ))}
      {folded.length > 0 && (
        <Dropdown
          triggerType="click"
          align="tr br"
          trigger={
            <button type="button" className="row-link">
              {i18n.t('More').toString()}
              <AiOutlineDown size={10} />
            </button>
          }
        >
          <Menu>
            {folded.map((l) => (
              <Menu.Item key={l.label} disabled={l.disabled} onClick={l.onClick}>
                <span className={classNames({ 'row-link-danger': l.danger && !l.disabled })}>{label(l)}</span>
              </Menu.Item>
            ))}
          </Menu>
        </Dropdown>
      )}
    </span>
  );
};
