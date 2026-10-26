import { connect } from 'dva';
import { Link } from 'dva/router';
import { Balloon } from '@alifd/next';
import React, { useEffect, useState } from 'react';
import { locationService } from '../../services/LocationService';
import { Translation } from '../../components/Translation';
import i18n from '../../i18n';
import type { SystemInfo, LoginUserInfo } from '@velaux/data';
import { LeftMenu, menuService } from '../../services/MenuService';

import './index.less';
import { checkPermission } from '../../types';
import { getConfigs } from '../../api/config';
import { Config, Menu, MenuTypes, PluginMeta } from '@velaux/data';
import { MdOutlineMonitorHeart } from 'react-icons/md';
import { AiFillStar, AiOutlineDown, AiOutlineStar, AiOutlineUp } from 'react-icons/ai';
import classNames from 'classnames';
import { readPref, writePref } from '../../utils/prefs';

interface Props {
  userInfo?: LoginUserInfo;
  systemInfo?: SystemInfo;
  pluginList: PluginMeta[];
  // collapsed is whether the sidebar shows icons only, so names show on hover.
  collapsed?: boolean;
  // filter narrows the menu to the items whose names contain it.
  filter?: string;
}

// The menus the user starred, and the sections they folded, by name.
const favoritesKey = 'menu-favorites';
const foldedKey = 'menu-folded';
const favoritesCatalog = 'My Favorites';

const LeftMenuModule = (props: Props) => {
  const path = locationService.getPathName();
  const [menus, setMenus] = useState<LeftMenu[]>();
  const [grafanaConfigs, setGrafanaConfigs] = useState<Config[]>();
  const [favorites, setFavorites] = useState<string[]>(readPref<string[]>(favoritesKey, []));
  const [folded, setFolded] = useState<string[]>(readPref<string[]>(foldedKey, []));
  useEffect(() => {
    if (checkPermission({ resource: 'config', action: 'list' }, '', props.userInfo)) {
      getConfigs('grafana').then((res) => {
        if (res) {
          setGrafanaConfigs(res.configs || []);
        }
      });
    }
  }, [props.userInfo]);

  useEffect(() => {
    menuService.resetPluginMenus();
    menuService.loadPluginMenus().then(() => {
      const menus = props.userInfo ? menuService.loadSidebarMenus(props.userInfo) : [];
      if (grafanaConfigs && grafanaConfigs.length > 0) {
        const grafanaLeftMenu: LeftMenu = { catalog: 'Grafana', menus: [] };
        grafanaConfigs.map((g) => {
          if (g.properties && g.properties['endpoint']) {
            grafanaLeftMenu.menus.push({
              name: g.name,
              workspace: 'extension',
              label: g.alias || g.name,
              to: '',
              href: g.properties['endpoint'],
              relatedRoute: [],
              type: MenuTypes.Workspace,
              icon: <MdOutlineMonitorHeart></MdOutlineMonitorHeart>,
            });
          }
        });
        if (grafanaLeftMenu.menus.length > 0) {
          menus.push(grafanaLeftMenu);
        }
      }
      setMenus(menus);
    });
  }, [props.userInfo, path, grafanaConfigs, props.pluginList]);

  if (!props.userInfo) {
    return <div />;
  }

  const toggleFavorite = (name: string) => {
    const next = favorites.includes(name) ? favorites.filter((f) => f !== name) : [...favorites, name];
    setFavorites(next);
    writePref(favoritesKey, next);
  };
  const toggleFolded = (catalog: string) => {
    const next = folded.includes(catalog) ? folded.filter((f) => f !== catalog) : [...folded, catalog];
    setFolded(next);
    writePref(foldedKey, next);
  };

  const keyword = (props.filter || '').trim().toLowerCase();
  const matches = (menu: Menu) =>
    !keyword ||
    i18n.t(menu.label).toString().toLowerCase().includes(keyword) ||
    menu.label.toLowerCase().includes(keyword);

  const renderItem = (childrenItem: Menu, section: string) => {
    const item = (
      <div className="menu-item-body">
        {childrenItem.icon}
        <span className={'menu-item-text'}>
          <Translation>{childrenItem.label}</Translation>
        </span>
      </div>
    );
    // hint names a collapsed item, and says when one is not available yet.
    const hint = (props.collapsed || childrenItem.comingSoon) && (
      <div className="menu-item-hint">
        {props.collapsed && (
          <div className="menu-item-hint-name">
            <Translation>{childrenItem.label}</Translation>
          </div>
        )}
        {childrenItem.comingSoon && (
          <div className="menu-item-hint-soon">
            <Translation>Coming Soon</Translation>
          </div>
        )}
      </div>
    );
    const withHint = (trigger: JSX.Element) =>
      hint ? (
        <Balloon.Tooltip align="r" trigger={trigger}>
          {hint}
        </Balloon.Tooltip>
      ) : (
        trigger
      );
    const key = `${section}/${childrenItem.name}`;
    if (childrenItem.comingSoon) {
      return (
        <li className="nav-item" key={key}>
          {withHint(
            <span className="menu-item menu-item-coming-soon" aria-disabled="true">
              {item}
            </span>
          )}
        </li>
      );
    }
    const favorite = favorites.includes(childrenItem.name);
    const star = !props.collapsed && (
      <span
        className={classNames('menu-item-star', { on: favorite })}
        title={i18n.t(favorite ? 'Remove from favorites' : 'Add to favorites').toString()}
        onClick={(e) => {
          e.preventDefault();
          e.stopPropagation();
          toggleFavorite(childrenItem.name);
        }}
      >
        {favorite ? <AiFillStar /> : <AiOutlineStar />}
      </span>
    );
    // The tooltip names a collapsed item, so the browser's title is only
    // given when it is not shown.
    const title = hint ? undefined : i18n.t(childrenItem.label).toString();
    const className = childrenItem.active ? 'menu-item-active' : 'menu-item';
    return (
      <li className="nav-item" key={key}>
        {childrenItem.href &&
          withHint(
            <a rel="noopener noreferrer" target="_blank" className={className} href={childrenItem.href} title={title}>
              {item}
              {star}
            </a>
          )}
        {childrenItem.to &&
          !childrenItem.href &&
          withHint(
            <Link to={childrenItem.to} className={className} title={title}>
              {item}
              {star}
            </Link>
          )}
      </li>
    );
  };

  const allMenus = (menus || []).flatMap((section) => section.menus);
  const favoriteMenus = favorites
    .map((name) => allMenus.find((m) => m.name === name && !m.comingSoon))
    .filter((m): m is Menu => !!m);
  const sections: LeftMenu[] = [
    ...(favoriteMenus.length > 0 ? [{ catalog: favoritesCatalog, menus: favoriteMenus }] : []),
    ...(menus || []),
  ];

  let fallBackCatalog = 0;
  const childrenSlider = sections.map((section) => {
    const catalog = section.catalog || `section-${fallBackCatalog++}`;
    const visible = section.menus.filter(matches);
    if (visible.length === 0) {
      return null;
    }
    // A filter opens every section it matches in.
    const open = props.collapsed || !!keyword || !folded.includes(catalog);
    const active = visible.some((m) => m.active);
    return (
      <li className="nav-container" key={catalog}>
        {section.catalog && !props.collapsed && (
          <div className={classNames('main-nav', { active: active && !open })} onClick={() => toggleFolded(catalog)}>
            {section.catalog === favoritesCatalog && <AiOutlineStar className="main-nav-icon" />}
            <span className="main-nav-text">
              <Translation>{section.catalog}</Translation>
            </span>
            {open ? <AiOutlineUp size={12} /> : <AiOutlineDown size={12} />}
          </div>
        )}
        {open && <ul className="sub-wrapper">{visible.map((m) => renderItem(m, catalog))}</ul>}
      </li>
    );
  });

  return (
    <div className="slide-wrapper">
      <ul className="ul-wrapper">{childrenSlider}</ul>
      {keyword && childrenSlider.every((c) => !c) && (
        <div className="menu-empty">
          <Translation>No matches</Translation>
        </div>
      )}
    </div>
  );
};

export default connect(
  (store: any) => {
    return { ...store.user, ...store.plugins };
  },
  undefined,
  undefined,
  { pure: false }
)(LeftMenuModule);
