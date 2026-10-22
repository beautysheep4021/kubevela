import React, { useEffect, useState } from 'react';
import classNames from 'classnames';
import { AiOutlineMenuFold, AiOutlineMenuUnfold, AiOutlineSearch } from 'react-icons/ai';

import i18n from '../../i18n';
import LeftMenu from '../LeftMenu';
import type { Customisation } from '../../services/CustomisationService';
import { customisationService } from '../../services/CustomisationService';
import { sidebarTheme } from '../../utils/theme';
import './index.less';

type Props = {
  // collapsed shows the menu as icons only.
  collapsed: boolean;
  onToggleCollapsed: () => void;
};

// Sidebar is the console's navigation column: a filter over the menu, the
// menu itself, and the switch that narrows it to icons.
export const Sidebar = (props: Props) => {
  const { collapsed, onToggleCollapsed } = props;
  const [filter, setFilter] = useState('');
  const [customisation, setCustomisation] = useState<Customisation>(customisationService.get());
  useEffect(() => customisationService.subscribe(setCustomisation), []);
  const theme = customisation.sidebarColor
    ? sidebarTheme(customisation.sidebarColor, customisation.accentColor)
    : sidebarTheme(undefined, customisation.accentColor);

  return (
    <div className={classNames('console-sidebar', { collapsed })} style={theme as React.CSSProperties}>
      {!collapsed && (
        <div className="sidebar-filter">
          <AiOutlineSearch size={14} />
          <input
            value={filter}
            placeholder={i18n.t('Search menu').toString()}
            onChange={(e) => setFilter(e.target.value)}
          />
        </div>
      )}
      <div className="sidebar-menu">
        <LeftMenu collapsed={collapsed} filter={collapsed ? '' : filter} />
      </div>
      <div
        className="sidebar-collapse"
        title={i18n.t(collapsed ? 'Expand' : 'Minimise').toString()}
        onClick={onToggleCollapsed}
      >
        {collapsed ? <AiOutlineMenuUnfold size={16} /> : <AiOutlineMenuFold size={16} />}
      </div>
    </div>
  );
};
