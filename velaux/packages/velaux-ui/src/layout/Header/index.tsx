import React, { Component } from 'react';
import './index.less';
import { Button, Dialog, Dropdown } from '@alifd/next';
import axios from 'axios';
import { connect } from 'dva';
import { Link, routerRedux } from 'dva/router';
import { AiOutlineCode, AiOutlineDown, AiOutlineHome, AiOutlineLogout, AiOutlineQuestionCircle } from 'react-icons/ai';

import logo from '../../assets/kubevela-logo-white.png';
import logoMark from '../../assets/KubeVela-01.svg';
import { loadCustomisation } from '../../api/customisation';
import type { Customisation } from '../../services/CustomisationService';
import { customisationService } from '../../services/CustomisationService';
import { If } from '../../components/If';
import Permission from '../../components/Permission';
import SwitchLanguage from '../../components/SwitchButton/index';
import { Translation } from '../../components/Translation';
import i18n from '../../i18n';
import type { AddonBaseStatus, Config, SystemInfo, LoginUserInfo } from '@velaux/data';
import { getData, setData } from '../../utils/cache';
import { locale } from '../../utils/locale';
import { getBrowserNameAndVersion } from '../../utils/utils';
import CloudShell from '../CloudShell';
import QuickSearch from '../../components/QuickSearch';
import { ProjectPicker } from '../../components/ProjectPicker';

import { LayoutMode, Workspace } from '@velaux/data';
import { Dispatch } from 'redux';
import { menuService } from '../../services/MenuService';
import classNames from 'classnames';

type Props = {
  dispatch: Dispatch;
  mode: LayoutMode;
  userInfo?: LoginUserInfo;
  systemInfo?: SystemInfo;
  show?: boolean;
  enabledAddons?: AddonBaseStatus[];
  currentWorkspace?: Workspace;
};

type State = {
  platformSetting: boolean;
  grafanaConfigs?: Config[];
  workspaces: Workspace[];
  customisation: Customisation;
};

const TelemetryDataCollectionKey = 'telemetryDataCollection';
const TelemetryDataCollectionServer = 'https://telemetry.kubevela.net/collecting';
@connect((store: any) => {
  return { ...store.user, ...store.cloudshell, ...store.addons };
})
class Header extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      platformSetting: false,
      workspaces: [],
      customisation: customisationService.get(),
    };
  }

  unsubscribe?: () => void;

  componentDidMount() {
    this.loadSystemInfo();
    this.loadUserInfo();
    this.loadEnabledAddons();
    this.unsubscribe = customisationService.subscribe((customisation) => this.setState({ customisation }));
    loadCustomisation();
  }

  componentWillUnmount() {
    this.unsubscribe?.();
  }

  loadWorkspaces = () => {
    const { userInfo } = this.props;
    menuService.loadPluginMenus().then(() => {
      this.setState({
        workspaces: menuService.loadWorkspaces(userInfo),
      });
    });
  };

  loadSystemInfo = () => {
    this.props.dispatch({
      type: 'user/getSystemInfo',
      callback: () => {
        this.telemetryDataCollection();
      },
    });
  };

  loadEnabledAddons = () => {
    this.props.dispatch({
      type: 'addons/getEnabledAddons',
      payload: {},
    });
  };

  telemetryDataCollection = async () => {
    const { systemInfo } = this.props;
    if (!getData(TelemetryDataCollectionKey) && systemInfo?.enableCollection) {
      try {
        axios
          .post(TelemetryDataCollectionServer, this.buildTelemetryData())
          .catch()
          .then(() => {
            this.setCache();
          });
      } catch {}
    }
  };

  buildTelemetryData = () => {
    const { systemInfo } = this.props;
    return {
      platformID: systemInfo?.platformID,
      installTime: systemInfo?.installTime,
      version: (systemInfo?.systemVersion?.velaVersion || '') + +'/' + (systemInfo?.systemVersion?.gitVersion || ''),
      clusterCount: systemInfo?.statisticInfo.clusterCount || '',
      appCount: systemInfo?.statisticInfo.appCount || '',
      enableAddonList: systemInfo?.statisticInfo.enableAddonList || {},
      componentDefinitionTopList: systemInfo?.statisticInfo.componentDefinitionTopList,
      traitDefinitionTopList: systemInfo?.statisticInfo.traitDefinitionTopList,
      workflowStepDefinitionTopList: systemInfo?.statisticInfo.workflowDefinitionTopList,
      policyDefinitionTopList: systemInfo?.statisticInfo.policyDefinitionTopList,
      browserInfo: {
        language: navigator.language,
        nameAndVersion: getBrowserNameAndVersion(),
        screenWidth: window.screen.width,
        screenHeight: window.screen.height,
      },
    };
  };

  setCache = () => {
    const now = new Date();
    now.setHours(now.getHours() + 24);
    setData(TelemetryDataCollectionKey, 'true', now);
  };

  loadUserInfo = () => {
    this.props.dispatch({
      type: 'user/getLoginUserInfo',
      callback: () => {
        this.loadWorkspaces();
      },
    });
  };

  onLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('refreshToken');
    this.props.dispatch(
      routerRedux.push({
        pathname: '/login',
      })
    );
  };

  checkEnabledAddon = (addonName: string) => {
    const { enabledAddons } = this.props;
    if (!enabledAddons) {
      return false;
    }
    const addonNames = enabledAddons.map((addon) => {
      return addon.name;
    });
    if (addonNames.includes(addonName)) {
      return true;
    }
    return false;
  };

  onOpenCloudShell = () => {
    if (!this.checkEnabledAddon('cloudshell')) {
      Dialog.alert({
        title: i18n.t('CloudShell feature is not enabled').toString(),
        content: i18n.t('You must enable the CloudShell addon').toString(),
        locale: locale().Dialog,
        footer: (
          <Button
            type="secondary"
            onClick={() => {
              this.props.dispatch(
                routerRedux.push({
                  pathname: '/addons/cloudshell',
                })
              );
            }}
          >
            <Translation>Go to enable</Translation>
          </Button>
        ),
      });
      return;
    }
    this.props.dispatch({
      type: 'cloudshell/open',
    });
  };

  render() {
    const { show, userInfo } = this.props;
    const admin = this.state.workspaces.find((ws) => ws.name === 'admin');
    // The admin screens open from the user's menu rather than the sidebar.
    const adminMenus =
      admin && userInfo ? menuService.loadMenus(admin, userInfo).flatMap((section) => section.menus) : [];
    const { customisation } = this.state;
    const wordmark = customisation.logoURL || logo;
    const userName = userInfo?.alias ? userInfo.alias : userInfo?.name;

    return (
      <div className="console-topnav">
        <Link to="/" className="topnav-brand" title={'Make shipping applications more enjoyable.'}>
          <img src={wordmark} className={classNames({ 'topnav-logo-plate': !!customisation.logoURL })} />
        </Link>
        <Link to="/" className="topnav-home">
          <AiOutlineHome size={16} />
          <span>
            <Translation>Console Home</Translation>
          </span>
        </Link>
        <div className="topnav-search">
          <QuickSearch userInfo={userInfo} />
        </div>
        <div className="topnav-end">
          <ProjectPicker />
          <Permission request={{ resource: 'cloudshell', action: 'create' }}>
            <span className="topnav-tool" title="Cloud Shell" onClick={this.onOpenCloudShell}>
              <AiOutlineCode size={18} />
            </span>
          </Permission>
          <a
            className="topnav-tool"
            title={i18n.t('Documentation').toString()}
            href="https://kubevela.io"
            target="_blank"
            rel="noopener noreferrer"
          >
            <AiOutlineQuestionCircle size={18} />
          </a>
          <span className="topnav-tool topnav-language">
            <SwitchLanguage />
          </span>
          <If condition={userInfo}>
            <Dropdown
              triggerType="hover"
              align="tr br"
              trigger={
                <div className="topnav-user" title={userName}>
                  <span className="topnav-avatar">{(userName || '?').slice(0, 1).toUpperCase()}</span>
                  <span className="topnav-user-name">{userName}</span>
                  <AiOutlineDown size={12} />
                </div>
              }
            >
              <div className="user-flyout">
                <div className="user-flyout-head">
                  <span className="topnav-avatar">{(userName || '?').slice(0, 1).toUpperCase()}</span>
                  <div>
                    <div className="user-flyout-name">{userName}</div>
                    {userInfo?.name && userInfo.name !== userName && (
                      <div className="user-flyout-sub">{userInfo.name}</div>
                    )}
                  </div>
                </div>
                {adminMenus.length > 0 && (
                  <div className="user-flyout-section">
                    <Translation>Admin</Translation>
                  </div>
                )}
                {adminMenus.map((menu) => (
                  <Link
                    key={menu.name}
                    to={menu.to}
                    className={classNames('user-flyout-item', { active: menu.active })}
                  >
                    {menu.icon}
                    <Translation>{menu.label}</Translation>
                  </Link>
                ))}
                <div className="user-flyout-divider" />
                <div className="user-flyout-item user-flyout-logout" onClick={this.onLogout}>
                  <AiOutlineLogout size={16} />
                  <Translation>Logout</Translation>
                </div>
              </div>
            </Dropdown>
          </If>
        </div>
        <If condition={show}>
          <CloudShell />
        </If>
        {(customisation.logoURL || customisation.iconURL) && (
          <a className="powered-by" href="https://kubevela.io" target="_blank" rel="noopener noreferrer">
            <Translation>Powered by</Translation>
            <img src={logoMark} alt="" />
            KubeVela
          </a>
        )}
      </div>
    );
  }
}

export default Header;
