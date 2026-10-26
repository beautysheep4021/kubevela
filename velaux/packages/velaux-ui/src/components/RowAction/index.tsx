import React from 'react';

import i18n from '../../i18n';
import './index.less';

type Props = {
  icon: React.ReactNode;
  label: string;
  danger?: boolean;
};

// RowAction is one of a row's actions, as a text link in the operations
// column, the way the cloud consoles list them. The icon is kept for callers
// that still pass one, but the label is what shows.
// It forwards its ref and any other props, so it can be a Dropdown's trigger.
export const RowAction = React.forwardRef<HTMLButtonElement, Props & React.ButtonHTMLAttributes<HTMLButtonElement>>(
  (props, ref) => {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const { icon, label: key, danger, className: extra, children, ...rest } = props;
    const label = i18n.t(key).toString();
    const className = `row-action ${danger ? 'danger' : ''} ${extra || ''}`;
    return (
      <button ref={ref} type="button" {...rest} className={className}>
        {label}
        {children}
      </button>
    );
  }
);

RowAction.displayName = 'RowAction';
