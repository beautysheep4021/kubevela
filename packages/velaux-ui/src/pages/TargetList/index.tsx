import { Button } from '@alifd/next';
import { projectChanged, scopedTo } from '../../utils/currentProject';
import { connect } from 'dva';
import React from 'react';

import { If } from '../../components/If';
import { ListToolbar, PageGuide, TablePager } from '../../components/Console';
import Permission from '../../components/Permission';
import { Translation } from '../../components/Translation';
import type { Cluster, Target, LoginUserInfo } from '@velaux/data';
import { Link } from 'dva/router';
import { AiOutlineAim, AiOutlineCluster, AiOutlineDeploymentUnit } from 'react-icons/ai';

import TableList from './components/List';
import TargetDialog from './components/TargetDialog';

type Props = {
  targets?: Target[];
  total?: number;
  currentProject?: { current: string; resolved: boolean };
  clusterList?: Cluster[];
  dispatch: ({}) => void;
  userInfo?: LoginUserInfo;
};

type State = {
  page: number;
  pageSize: number;
  query: string;
  showAddTarget: boolean;
  editTargetName: string;
  visibleDelivery: boolean;
  isEdit: boolean;
  targetItem?: Target;
};

@connect((store: any) => {
  return { ...store.target, ...store.clusters, ...store.user, currentProject: store.currentProject };
})
class TargetList extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      query: '',
      page: 1,
      pageSize: 10,
      showAddTarget: false,
      editTargetName: '',
      visibleDelivery: false,
      isEdit: false,
    };
  }

  componentDidMount() {
    this.getTargetList();
    this.getClusterList();
  }

  componentDidUpdate(prev: Props) {
    if (projectChanged(prev.currentProject, this.props.currentProject)) {
      this.getTargetList();
    }
  }

  // getTargetList lists the picked project's, or every project's for all of them.
  getTargetList = async () => {
    if (!this.props.currentProject?.resolved) {
      return;
    }
    const { page, pageSize, query } = this.state;
    this.props.dispatch({
      type: 'target/listTargets',
      payload: {
        query,
        page,
        pageSize,
        project: this.props.currentProject.current,
      },
    });
  };

  getClusterList = async () => {
    this.props.dispatch({
      type: 'clusters/getClusterList',
    });
  };

  updateTargetList = () => {
    this.setState(
      {
        query: '',
        page: 1,
      },
      () => {
        this.getTargetList();
      }
    );
  };

  changeISEdit = (isEdit: boolean, record: Target) => {
    this.setState({
      isEdit,
      visibleDelivery: true,
      targetItem: record,
    });
  };

  onClose = () => {
    this.setState({ visibleDelivery: false, isEdit: false });
  };

  onOk = () => {
    this.getTargetList();
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
        this.getTargetList();
      }
    );
  };

  // search finds targets by name from the first page.
  search = (query: string) => {
    this.setState({ query, page: 1 }, () => {
      this.getTargetList();
    });
  };

  render() {
    const { clusterList, targets, total } = this.props;
    const { visibleDelivery, isEdit, targetItem } = this.state;
    return (
      <div>
        <PageGuide
          id="targets"
          title="Targets"
          description="Define the targets that applications would deliver to"
          steps={[
            {
              icon: <AiOutlineCluster />,
              title: 'Prepare a cluster',
              description: <Translation>Connect the cluster the target points at</Translation>,
            },
            {
              icon: <AiOutlineAim />,
              title: 'Create a target',
              description: <Translation>Name a cluster and a namespace in it, for a project</Translation>,
            },
            {
              icon: <AiOutlineDeploymentUnit />,
              title: 'Use it in applications',
              description: (
                <Translation>Add it to an environment, and the applications there deliver to it</Translation>
              ),
            },
          ]}
          extra={
            <div className="page-guide-links">
              <Link to="/clusters">
                <Translation>Clusters</Translation>
              </Link>
              <Link to="/envs">
                <Translation>Environments</Translation>
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
              search={{ value: this.state.query, placeholder: 'Search by name', onSearch: this.search }}
              onRefresh={this.getTargetList}
              actions={
                <Permission key={'new-target'} request={{ resource: 'target:*', action: 'create' }} project={''}>
                  <Button
                    type="primary"
                    onClick={() => {
                      this.setState({ visibleDelivery: true, targetItem: undefined });
                    }}
                  >
                    <Translation>New Target</Translation>
                  </Button>
                </Permission>
              }
            />
          </div>

          <TableList
            list={scopedTo<Target>(targets, this.props.currentProject, (t: Target) => t.project?.name)}
            updateTargetList={this.updateTargetList}
            changeISEdit={(is: boolean, record: Target) => {
              this.changeISEdit(is, record);
            }}
          />

          <TablePager
            total={total || 0}
            page={this.state.page}
            pageSize={this.state.pageSize}
            onChange={this.handleChange}
          />
        </div>

        <If condition={visibleDelivery}>
          <TargetDialog
            visible={visibleDelivery}
            clusterList={clusterList || []}
            isEdit={isEdit}
            targetItem={targetItem}
            onClose={this.onClose}
            onOK={this.onOk}
          />
        </If>
      </div>
    );
  }
}

export default TargetList;
