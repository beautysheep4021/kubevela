import React, { useEffect, useState } from 'react';
import { Button } from '@alifd/next';
import classNames from 'classnames';
import { AiOutlineAppstore, AiOutlineReload, AiOutlineSearch, AiOutlineUnorderedList } from 'react-icons/ai';

import i18n from '../../i18n';
import { Translation } from '../Translation';

export type ViewMode = 'table' | 'card';

type Props = {
  // search is the name filter's value and placeholder; changes come through
  // onSearch on Enter, or as typed when live is set.
  search?: { value?: string; placeholder: string; live?: boolean; onSearch: (value: string) => void };
  // filters sit after the search, a page's own selects.
  filters?: React.ReactNode;
  onRefresh?: () => void;
  refreshing?: boolean;
  // actions sit at the right, the page's primary action last.
  actions?: React.ReactNode;
  view?: { value: ViewMode; onChange: (mode: ViewMode) => void };
};

// ListToolbar is the bar above a resource list: find on the left; refresh,
// the page's actions and the view switch on the right.
export const ListToolbar = (props: Props) => {
  const { search, filters, onRefresh, refreshing, actions, view } = props;
  const [value, setValue] = useState(search?.value || '');
  useEffect(() => setValue(search?.value || ''), [search?.value]);
  return (
    <div className="list-toolbar">
      <div className="list-toolbar-start">
        {search && (
          <label className="list-toolbar-search">
            <AiOutlineSearch />
            <input
              value={value}
              placeholder={i18n.t(search.placeholder).toString()}
              onChange={(e) => {
                setValue(e.target.value);
                if (search.live) {
                  search.onSearch(e.target.value);
                }
              }}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  search.onSearch(value);
                }
              }}
            />
          </label>
        )}
        {filters}
      </div>
      <div className="list-toolbar-end">
        {onRefresh && (
          <Button
            className="list-toolbar-icon"
            title={i18n.t('Refresh').toString()}
            aria-label={i18n.t('Refresh').toString()}
            onClick={onRefresh}
          >
            <AiOutlineReload className={classNames({ spinning: refreshing })} />
          </Button>
        )}
        {actions}
        {view && (
          <div className="list-toolbar-view" role="group" aria-label={i18n.t('View').toString()}>
            <span
              className={classNames({ active: view.value === 'table' })}
              title={i18n.t('Table').toString()}
              onClick={() => view.onChange('table')}
            >
              <AiOutlineUnorderedList />
            </span>
            <span
              className={classNames({ active: view.value === 'card' })}
              title={i18n.t('Card').toString()}
              onClick={() => view.onChange('card')}
            >
              <AiOutlineAppstore />
            </span>
          </div>
        )}
      </div>
    </div>
  );
};

// ToolbarButton is a toolbar action, primary for the page's main one.
export const ToolbarButton = (props: { label: string; primary?: boolean; onClick: () => void; disabled?: boolean }) => (
  <Button type={props.primary ? 'primary' : 'normal'} onClick={props.onClick} disabled={props.disabled}>
    <Translation>{props.label}</Translation>
  </Button>
);
