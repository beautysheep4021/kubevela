import { Balloon, Dialog, Message, Table } from '@alifd/next';
import { connect } from 'dva';

import Empty from '../Empty';

import { Link, routerRedux } from 'dva/router';
import React, { Component } from 'react';

import './index.less';
import type { Dispatch } from 'redux';

import { compareApplication, rollbackWithRevision } from '../../../../api/application';
import { ApplicationDiff } from '../../../../components/ApplicationDiff';
import { If } from '../../../../components/If';
import { RowLinks } from '../../../../components/Console';
import { StatusBadge } from '../../../../components/StatusBadge';
import Item from '../../../../components/Item';
import { Translation } from '../../../../components/Translation';
import type {
  ApplicationCompareRequest,
  ApplicationCompareResponse,
  ApplicationDetail,
  ApplicationRevision,
  ApplicationRollbackResponse,
  LoginUserInfo,
  NameAlias,
} from '@velaux/data';
import { checkPermission } from '../../../../utils/permission';
import { momentDate, showAlias } from '../../../../utils/common';
import { locale } from '../../../../utils/locale';
import { statusList, statusTones } from '../../constants';

type Props = {
  list: ApplicationRevision[];
  getRevisionList: () => void;
  applicationDetail?: ApplicationDetail;
  onShowAppModel: (record: ApplicationRevision) => void;
  dispatch: Dispatch<any>;
  userInfo?: LoginUserInfo;
};

type State = {
  compare?: ApplicationCompareResponse;
  revision?: ApplicationRevision;
  visibleApplicationDiff: boolean;
  diffMode: 'latest' | 'cluster';
};

@connect((store: any) => {
  return { userInfo: store.user.userInfo };
})
class TableList extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      diffMode: 'latest',
      visibleApplicationDiff: false,
    };
  }

  onRollback = (record: ApplicationRevision, ok?: boolean) => {
    const { applicationDetail, dispatch } = this.props;
    if (record.status === 'terminated' && !ok) {
      Dialog.confirm({
        type: 'confirm',
        content: (
          <Translation>This revision status is terminated, are you sure to rollback to this revision?</Translation>
        ),
        onOk: () => {
          this.onRollback(record, true);
        },
        locale: locale().Dialog,
      });
      return;
    }
    if (applicationDetail) {
      rollbackWithRevision({ appName: applicationDetail?.name, revision: record.version }).then(
        (res: ApplicationRollbackResponse) => {
          if (res) {
            Message.success('Application rollback successfully');
            if (res.record && res.record.name && dispatch) {
              dispatch(
                routerRedux.push(
                  `/applications/${applicationDetail.name}/envbinding/${record.envName}/workflow/records/${res.record.name}`
                )
              );
            }
          }
        }
      );
    }
  };

  loadChanges = (revision: ApplicationRevision, mode: 'latest' | 'cluster') => {
    const { applicationDetail } = this.props;
    if (!revision || !applicationDetail) {
      this.setState({ compare: undefined });
      return;
    }
    let params: ApplicationCompareRequest = {
      compareRevisionWithLatest: { revision: revision.version },
    };
    if (mode === 'cluster') {
      params = {
        compareRevisionWithRunning: { revision: revision.version },
      };
    }
    compareApplication(applicationDetail?.name, params).then((res: ApplicationCompareResponse) => {
      this.setState({
        revision: revision,
        compare: res,
        visibleApplicationDiff: true,
        diffMode: mode,
      });
    });
  };

  getColumns = () => {
    const { applicationDetail } = this.props;
    return [
      {
        key: 'version',
        title: <Translation>Revision</Translation>,
        dataIndex: 'version',
        cell: (v: string, i: number, record: ApplicationRevision) => {
          if (record.codeInfo) {
            return (
              <Balloon style={{ width: '220px' }} trigger={<a>{v}</a>}>
                <Item
                  label="Commit"
                  value={<a title={record.codeInfo.commit}>{record.codeInfo.commit?.slice(0, 7)}</a>}
                />
                <Item label="Branch" value={record.codeInfo.branch} />
                <Item label="User" value={record.codeInfo.user} />
              </Balloon>
            );
          }
          return <span>{v}</span>;
        },
      },
      {
        key: 'deployUser',
        title: <Translation>Deploy User</Translation>,
        dataIndex: 'deployUser',
        cell: (v: NameAlias) => {
          return showAlias(v.name, v.alias);
        },
      },
      {
        key: 'status',
        title: <Translation>Status</Translation>,
        dataIndex: 'status',
        cell: (v: string, i: number, record: ApplicationRevision) => {
          const findObj = statusList.find((item) => item.value === v);
          if (findObj) {
            const show = <StatusBadge tone={statusTones[v] || 'neutral'} label={findObj.label} />;
            if (record.reason) {
              return <Balloon trigger={show}>{record.reason}</Balloon>;
            }
            return show;
          }
          return '';
        },
      },
      {
        key: 'triggerType',
        title: <Translation>Trigger Type</Translation>,
        dataIndex: 'triggerType',
      },
      {
        key: 'createTime',
        title: <Translation>Publish Time</Translation>,
        dataIndex: 'createTime',
        cell: (v: string) => {
          return <span>{v && momentDate(v)}</span>;
        },
      },
      {
        key: 'envName',
        title: <Translation>Environment</Translation>,
        dataIndex: 'envName',
        cell: (v: string) => {
          return (
            <span>
              <Link to={`/applications/${applicationDetail && applicationDetail.name}/envbinding/${v}/status`}>
                {v}
              </Link>
            </span>
          );
        },
      },
      {
        key: 'operation',
        title: <Translation>Actions</Translation>,
        dataIndex: 'operation',
        width: 260,
        lock: 'right',
        cell: (v: string, i: number, record: ApplicationRevision) => {
          const project = applicationDetail?.project?.name;
          const app = `project:${project}/application:${applicationDetail?.name}`;
          const { userInfo } = this.props;
          return (
            <RowLinks
              visible={2}
              links={[
                {
                  label: 'Detail',
                  hidden: !checkPermission({ resource: app, action: 'detail' }, project, userInfo),
                  onClick: () => this.props.onShowAppModel(record),
                },
                {
                  label: 'Rollback',
                  hidden:
                    !(record.status === 'complete' || record.status == 'terminated') ||
                    !checkPermission(
                      { resource: `${app}/revision:${record.version}`, action: 'rollback' },
                      project,
                      userInfo
                    ),
                  onClick: () => this.onRollback(record),
                },
                // The diffs' names are long, so they always sit under More.
                {
                  label: 'Diff With Deployed Application',
                  more: true,
                  onClick: () => this.loadChanges(record, 'cluster'),
                },
                {
                  label: 'Diff With Latest Configuration',
                  more: true,
                  onClick: () => this.loadChanges(record, 'latest'),
                },
              ]}
            />
          );
        },
      },
    ];
  };

  render() {
    const { Column } = Table;
    const columns = this.getColumns();
    const { list } = this.props;
    const { visibleApplicationDiff, compare, diffMode, revision } = this.state;
    const baseName = 'Current Revision';
    let targetName = 'Latest Application Configuration';
    if (diffMode == 'cluster') {
      targetName = 'Deployed Application';
    }
    return (
      <div className="table-version-list">
        <Table
          locale={locale().Table}
          primaryKey={'version'}
          className="customTable"
          rowHeight={40}
          dataSource={list}
          loading={false}
          emptyContent={<Empty />}
        >
          {columns.map((col) => (
            <Column {...col} key={col.key} align={'left'} />
          ))}
        </Table>
        <If condition={visibleApplicationDiff}>
          {compare && (
            <ApplicationDiff
              onClose={() => {
                this.setState({
                  visibleApplicationDiff: false,
                  compare: undefined,
                  revision: undefined,
                });
              }}
              rollback2Revision={
                diffMode === 'cluster' && revision
                  ? () => {
                      this.onRollback(revision);
                    }
                  : undefined
              }
              baseName={baseName}
              targetName={targetName}
              compare={compare}
            />
          )}
        </If>
      </div>
    );
  }
}

export default TableList;
