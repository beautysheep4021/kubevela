import React, { useEffect, useState } from 'react';
import classNames from 'classnames';
import { Link } from 'dva/router';
import { AiOutlineClose } from 'react-icons/ai';

import i18n from '../../i18n';
import { locationService } from '../../services/LocationService';
import { menuService } from '../../services/MenuService';
import { readPref, writePref } from '../../utils/prefs';
import type { Tab } from './tabs';
import { closeTab, openTab, tabFor } from './tabs';
import './index.less';

const tabsKey = 'recent-tabs';
const maxTabs = 10;

// tabAt names the page at a path from the menu that holds it.
const tabAt = (path: string) => tabFor(path, menuService.findMenu(path));

// RecentTabs keeps the pages the user opened as tabs above the content, as
// the cloud consoles do, so moving between them takes one click.
export const RecentTabs = () => {
  const [tabs, setTabs] = useState<Tab[]>(readPref<Tab[]>(tabsKey, []));
  const [path, setPath] = useState(locationService.getPathName());

  const save = (next: Tab[]) => {
    setTabs(next);
    writePref(tabsKey, next);
  };

  useEffect(() => {
    const history = locationService.getHistory();
    const visit = (pathname: string) => {
      setPath(pathname);
      menuService.loadPluginMenus().then(() => {
        const tab = tabAt(pathname);
        if (!tab) {
          return;
        }
        setTabs((current) => {
          const next = openTab(current, tab, maxTabs);
          writePref(tabsKey, next);
          return next;
        });
      });
    };
    visit(history.location.pathname);
    return history.listen((location) => visit(location.pathname));
  }, []);

  if (tabs.length === 0) {
    return null;
  }

  const current = tabAt(path)?.key;
  const close = (tab: Tab) => {
    const closed = closeTab(tabs, tab.key, current);
    save(closed.tabs);
    if (closed.goTo) {
      locationService.push(closed.goTo);
    }
  };

  return (
    <div className="recent-tabs" role="tablist">
      {tabs.map((tab) => (
        <Link
          key={tab.key}
          to={tab.path}
          role="tab"
          aria-selected={tab.key === current}
          className={classNames('recent-tab', { active: tab.key === current })}
          title={tab.detail ? `${i18n.t(tab.label)} / ${tab.detail}` : i18n.t(tab.label).toString()}
        >
          <span className="recent-tab-label">
            {i18n.t(tab.label).toString()}
            {tab.detail && <span className="recent-tab-detail">{tab.detail}</span>}
          </span>
          <span
            className="recent-tab-close"
            aria-label={i18n.t('Close').toString()}
            onClick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              close(tab);
            }}
          >
            <AiOutlineClose size={10} />
          </span>
        </Link>
      ))}
    </div>
  );
};
