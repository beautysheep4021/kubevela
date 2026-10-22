import { Button } from '@alifd/next';
import { projectChanged, scopedTo } from '../../utils/currentProject';
import { connect } from 'dva';
import React from 'react';

import { If } from '../../components/If';
import { ListToolbar, PageGuide, TablePager } from '../../components/Console';
import Permission from '../../components/Permission';
import { Translation } from '../../components/Translation';
import type { Env, LoginUserInfo } from '@velaux/data';
import { Link } from 'dva/router';
import { AiOutlineAim, AiOutlineCluster, AiOutlineDeploymentUnit } from 'react-icons/ai';

import EnvDialog from './components/EnvDialog';
import TableList from './components/List';

type Props = {
  envTotal?: number;
  envs: Env[];
  currentProject?: { current: string; resolved: boolean };
  dispatch: ({}) => void;
  userInfo?: LoginUserInfo;
};

type State = {
  page: number;
  pageSize: number;
  showAddTarget: boolean;
  editTargetName: string;
  visibleEnvDialog: boolean;
  isEdit: boolean;
  envItem?: Env;
  // search filters the loaded page by name, alias or namespace.
  search: string;
};

@connect((store: any) => {
  return { ...store.target, ...store.application, ...store.env, ...store.user, currentProject: store.currentProject };
})
class EnvList extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      page: 1,
      pageSize: 10,
      showAddTarget: false,
      editTargetName: '',
      visibleEnvDialog: false,
      isEdit: false,
      search: '',
    };
  }

  componentDidMount() {
    this.getEnvList();
  }

  componentDidUpdate(prev: Props) {
    if (projectChanged(prev.currentProject, this.props.currentProject)) {
      this.getEnvList();
    }
  }

  // getEnvList lists the picked project's, or every project's for all of them.
  getEnvList = async () => {
    if (!this.props.currentProject?.resolved) {
      return;
    }
    const { page, pageSize } = this.state;
    this.props.dispatch({
      type: 'env/listEnvs',
      payload: {
        page,
        pageSize,
        project: this.props.currentProject.current,
      },
    });
  };

  updateEnvList = () => {
    this.setState(
      {
        page: 1,
      },
      () => {
        this.getEnvList();
      }
    );
  };

  changeISEdit = (isEdit: boolean, record: Env) => {
    this.setState({
      isEdit,
      visibleEnvDialog: true,
      envItem: record,
    });
  };

  onClose = () => {
    this.setState({ visibleEnvDialog: false, isEdit: false });
  };

  onOk = () => {
    this.getEnvList();
    this.setState({
      isEdit: false,
    });
  };

  handleChange = (page: number, pageSize: number) => {
    this.setState(
      {
        page,
        pageSize,
      },
      () => {
        this.getEnvList();
      }
    );
  };

  render() {
    const { envTotal, envs, userInfo } = this.props;
    const { visibleEnvDialog, isEdit, envItem, search } = this.state;
    const keyword = search.trim().toLowerCase();
    const list = scopedTo(envs, this.props.currentProject, (e: Env) => e.project?.name)?.filter(
      (e: Env) => !keyword || [e.name, e.alias, e.namespace].some((v) => (v || '').toLowerCase().includes(keyword))
    );
    return (
      <div>
        <PageGuide
          id="environments"
          title="Environments"
          description="Set up the Environments for your Application based on Target sources"
          steps={[
            {
              icon: <AiOutlineCluster />,
              title: 'Prepare targets',
              description: (
                <Translation>Connect clusters, and create the targets an environment deploys to</Translation>
              ),
            },
            {
              icon: <AiOutlineAim />,
              title: 'Create an environment',
              description: <Translation>Give a project an environment, with its namespace and targets</Translation>,
            },
            {
              icon: <AiOutlineDeploymentUnit />,
              title: 'Use it in applications',
              description: <Translation>Bind applications to it, and each deploys there on its own</Translation>,
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
              <Link to="/applications">
                <Translation>Applications</Translation>
              </Link>
            </div>
          }
        />

        <div className="console-panel">
          <div className="console-panel-body" style={{ paddingBottom: 4 }}>
            <ListToolbar
              search={{
                value: search,
                placeholder: 'Search by name or namespace',
                live: true,
                onSearch: (value: string) => this.setState({ search: value }),
              }}
              onRefresh={this.getEnvList}
              actions={
                <Permission
                  key={'new-env'}
                  request={{ resource: 'project:?/environment:*', action: 'create' }}
                  project={'?'}
                >
                  <Button
                    type="primary"
                    onClick={() => {
                      this.setState({ visibleEnvDialog: true, envItem: undefined });
                    }}
                  >
                    <Translation>New Environment</Translation>
                  </Button>
                </Permission>
              }
            />
          </div>

          <TableList
            list={list}
            updateEnvList={this.updateEnvList}
            userInfo={userInfo}
            changeISEdit={(is: boolean, record: Env) => {
              this.changeISEdit(is, record);
            }}
          />

          <TablePager
            total={envTotal || 0}
            page={this.state.page}
            pageSize={this.state.pageSize}
            onChange={this.handleChange}
          />
        </div>

        <If condition={visibleEnvDialog}>
          <EnvDialog
            visible={visibleEnvDialog}
            projects={userInfo?.projects || []}
            isEdit={isEdit}
            userInfo={userInfo}
            envItem={envItem}
            onClose={this.onClose}
            onOK={this.onOk}
          />
        </If>
      </div>
    );
  }
}

export default EnvList;
