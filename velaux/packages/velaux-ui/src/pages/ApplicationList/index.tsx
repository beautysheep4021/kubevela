import { Message, Loading, Button } from '@alifd/next';
import { connect } from 'dva';
import React, { Component } from 'react';
import type { HealthFilter } from './components/AppStatus/health';
import { byHealth, healthCounts } from './components/AppStatus/health';
import { HealthChips } from './components/HealthChips';
import { visibleLabels } from '../../utils/appMeta';

import { projectChanged, scopedTo } from '../../utils/currentProject';
import { deleteApplication } from '../../api/application';
import { If } from '../../components/If';
import { PageGuide } from '../../components/Console';
import Permission from '../../components/Permission';
import { Translation } from '../../components/Translation';
import type { ApplicationBase, Env, LoginUserInfo } from '@velaux/data';

import { Link } from 'dva/router';
import { AiOutlineAppstoreAdd, AiOutlineCluster, AiOutlineDeploymentUnit } from 'react-icons/ai';
import CardContend from './components/CardContent';
import EditAppDialog from './components/EditAppDialog';
import SelectSearch from './components/SelectSearch';

type Props = {
  dispatch: ({}) => {};
  applicationList?: ApplicationBase[];
  targets?: [];
  envs?: [];
  history: any;
  userInfo?: LoginUserInfo;
  currentProject?: { current: string; resolved: boolean };
};

export type ShowMode = 'table' | 'card' | string | null;

type State = {
  isLoading: boolean;
  showEditApplication: boolean;
  editItem?: ApplicationBase;
  labelValue: string[];
  showMode: ShowMode;
  health: HealthFilter;
};

@connect((store: any) => {
  return {
    ...store.application,
    ...store.target,
    ...store.clusters,
    ...store.env,
    ...store.user,
    currentProject: store.currentProject,
  };
})
class Application extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    let mode: ShowMode = localStorage.getItem('application-list-mode');
    if (mode != 'table' && mode != 'card') {
      mode = 'table';
    }
    this.state = {
      labelValue: [],
      isLoading: false,
      showEditApplication: false,
      showMode: mode,
      health: 'all',
    };
  }

  componentDidMount() {
    this.getApplications({});
    this.getEnvs();
  }

  componentDidUpdate(prev: Props) {
    if (projectChanged(prev.currentProject, this.props.currentProject)) {
      this.getApplications({});
      this.getEnvs();
    }
  }

  // getApplications lists the picked project's applications, or every project's
  // for all of them, and not before the project is known.
  getApplications = async (params: any) => {
    if (!this.props.currentProject?.resolved) {
      return;
    }
    this.setState({ isLoading: true });
    this.props.dispatch({
      type: 'application/getApplicationList',
      // An addon's applications are listed on the Addons page instead.
      payload: { ...params, project: this.props.currentProject.current, withStatus: true, addons: 'exclude' },
      callback: () => {
        this.setState({
          isLoading: false,
        });
      },
    });
  };

  // getEnvs lists the environments the Environment filter offers, the picked
  // project's.
  getEnvs = async () => {
    if (!this.props.currentProject?.resolved) {
      return;
    }
    this.props.dispatch({
      type: 'env/listEnvs',
      payload: { project: this.props.currentProject.current },
    });
  };

  setLabelValue = async (labels: string[]) => {
    this.setState({
      labelValue: labels,
    });
  };

  onDeleteApp = (name: string) => {
    deleteApplication({ name: name }).then((re) => {
      if (re) {
        Message.success('Application deleted successfully');
        this.getApplications({});
      }
    });
  };

  // addApplication opens the create page.
  addApplication = () => {
    this.props.history.push('/applications/new');
  };

  closeEditAppDialog = () => {
    this.setState({
      showEditApplication: false,
    });
    this.getApplications({});
  };

  editAppPlan = (editItem: ApplicationBase) => {
    this.setState({
      editItem,
      showEditApplication: true,
    });
  };

  clickLabelFilter = (label: string) => {
    let { labelValue } = this.state;
    let existIndex = -1;
    labelValue.map((key, index) => {
      if (key == label) {
        existIndex = index;
        return;
      }
    });
    if (existIndex == -1) {
      labelValue.push(label);
    } else {
      labelValue = labelValue.splice(existIndex, existIndex);
    }
    this.setState({
      labelValue,
    });
    this.getApplications({ labels: labelValue.join(',') });
  };

  render() {
    const { dispatch, envs } = this.props;
    const applicationList = scopedTo(this.props.applicationList, this.props.currentProject, (a) => a.project?.name);
    const { isLoading, showEditApplication, editItem, labelValue, showMode } = this.state;
    let appLabels: string[] = [];
    applicationList?.map((app) => {
      app.labels &&
        Object.keys(app.labels).map((key: string) => {
          if (visibleLabels(app.labels).includes(key)) {
            if (app.labels) {
              appLabels.push(key + '=' + app.labels[key]);
            }
          }
        });
    });
    return (
      <div>
        <PageGuide
          id="applications"
          title="Applications"
          description="Deploy and manage all your applications"
          steps={[
            {
              icon: <AiOutlineCluster />,
              title: 'Prepare',
              description: <Translation>Connect clusters, then set up delivery targets and environments</Translation>,
            },
            {
              icon: <AiOutlineAppstoreAdd />,
              title: 'Create an application',
              description: <Translation>Name it, bind its environments, then add its components</Translation>,
            },
            {
              icon: <AiOutlineDeploymentUnit />,
              title: 'Deploy and observe',
              description: <Translation>Run its workflow, then watch its status, resources and revisions</Translation>,
            },
          ]}
          extra={
            <div className="page-guide-links">
              <Link to="/clusters">
                <Translation>Clusters</Translation>
              </Link>
              <Link to="/targets">
                <Translation>Targets</Translation>
              </Link>
              <Link to="/envs">
                <Translation>Environments</Translation>
              </Link>
              <Link to="/shared-workflows">
                <Translation>Workflows</Translation>
              </Link>
            </div>
          }
        />

        <div className="console-panel">
          <SelectSearch
            appLabels={appLabels}
            dispatch={dispatch}
            setLabelValue={this.setLabelValue}
            labelValue={labelValue}
            envs={scopedTo(envs, this.props.currentProject, (env: Env) => env.project?.name)}
            showMode={showMode}
            setMode={(mode: ShowMode) => {
              this.setState({ showMode: mode });
              if (mode) {
                localStorage.setItem('application-list-mode', mode);
              }
            }}
            getApplications={(params: any) => {
              this.getApplications(params);
            }}
            actions={
              <Permission request={{ resource: 'project:?/application:*', action: 'create' }} project={'?'}>
                <Button type="primary" onClick={this.addApplication}>
                  <Translation>New Application</Translation>
                </Button>
              </Permission>
            }
          />
          <HealthChips
            counts={healthCounts(applicationList || [])}
            value={this.state.health}
            onChange={(health: HealthFilter) => this.setState({ health })}
          />
          <Loading visible={isLoading} fullScreen>
            <CardContend
              applications={byHealth(applicationList || [], this.state.health)}
              editAppPlan={(item: ApplicationBase) => {
                this.editAppPlan(item);
              }}
              clickLabelFilter={this.clickLabelFilter}
              showMode={showMode}
              deleteAppPlan={this.onDeleteApp}
              setVisible={(visible) => {
                if (visible) {
                  this.addApplication();
                }
              }}
            />
          </Loading>
        </div>
        <If condition={showEditApplication && editItem}>
          <EditAppDialog editItem={editItem} onOK={this.closeEditAppDialog} onClose={this.closeEditAppDialog} />
        </If>
      </div>
    );
  }
}

export default Application;
