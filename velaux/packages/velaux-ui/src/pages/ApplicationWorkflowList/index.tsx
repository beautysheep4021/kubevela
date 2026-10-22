import { Dialog, Message, Table } from '@alifd/next';
import { connect } from 'dva';
import { Link, routerRedux } from 'dva/router';
import React from 'react';
import type { Dispatch } from 'redux';

import { deleteWorkflow } from '../../api/workflows';
import { ListToolbar, RowLinks, TablePager, pageOf } from '../../components/Console';
import { Translation } from '../../components/Translation';
import i18n from '../../i18n';
import type { ApplicationDetail, EnvBinding, LoginUserInfo, Workflow } from '@velaux/data';
import { checkPermission } from '../../utils/permission';
import { momentDate } from '../../utils/common';
import { locale } from '../../utils/locale';

type Props = {
  revisions: [];
  applicationDetail?: ApplicationDetail;
  envbinding?: EnvBinding[];
  workflows: Workflow[];
  dispatch: Dispatch<any>;
  userInfo?: LoginUserInfo;
  match: {
    params: {
      appName: string;
    };
  };
};

type State = {
  page: number;
  pageSize: number;
};

@connect((store: any) => {
  return { ...store.application, userInfo: store.user.userInfo };
})
class ApplicationWorkflowList extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { page: 1, pageSize: 10 };
  }

  componentDidMount() {}

  loadApplicationWorkflows = async () => {
    const {
      params: { appName },
    } = this.props.match;
    this.props.dispatch({
      type: 'application/getApplicationWorkflows',
      payload: { appName: appName },
    });
  };

  onDeleteWorkflow = (name: string) => {
    const { applicationDetail } = this.props;
    if (applicationDetail) {
      Dialog.confirm({
        type: 'confirm',
        content: <Translation>Unrecoverable after deletion, are you sure to delete it?</Translation>,
        onOk: () => {
          deleteWorkflow({
            appName: applicationDetail.name,
            name: name,
          }).then((res) => {
            if (res) {
              Message.success(i18n.t('Workflow removed successfully'));
              this.loadApplicationWorkflows();
            }
          });
        },
        locale: locale().Dialog,
      });
    }
  };

  render() {
    const { applicationDetail, workflows, userInfo } = this.props;
    const projectName = applicationDetail?.project?.name;
    const list = workflows || [];
    const { pageSize } = this.state;
    const page = Math.min(this.state.page, Math.max(1, Math.ceil(list.length / pageSize)));
    const studio = (w: Workflow) =>
      `/applications/${applicationDetail?.name}/envbinding/${w.envName}/workflow/${w.name}/studio`;
    const may = (w: Workflow, action: string) =>
      checkPermission(
        { resource: `project:${projectName}/application:${applicationDetail?.name}/workflow:${w.name}`, action },
        projectName,
        userInfo
      );
    return (
      <div className="console-panel app-tab-panel">
        <ListToolbar
          filters={
            <span className="app-tab-hint">
              <Translation>The workflow each environment deploys with. Open one to edit its steps.</Translation>
            </span>
          }
          onRefresh={this.loadApplicationWorkflows}
        />
        <Table dataSource={pageOf(list, page, pageSize)} primaryKey="name" locale={locale().Table}>
          <Table.Column
            dataIndex="name"
            title={i18n.t('Name').toString()}
            cell={(v: string, i: number, w: Workflow) => {
              if (applicationDetail?.readOnly) {
                return v;
              }
              return <Link to={studio(w)}>{v}</Link>;
            }}
          />
          <Table.Column
            dataIndex="envName"
            title={i18n.t('Environment').toString()}
            cell={(v: string) => {
              return <Link to={`/applications/${applicationDetail?.name}/envbinding/${v}/workflow`}>{v}</Link>;
            }}
          />
          <Table.Column dataIndex="mode" title={i18n.t('Mode').toString()} />
          <Table.Column
            dataIndex="steps"
            title={i18n.t('Step Count').toString()}
            cell={(steps: []) => {
              return steps ? steps.length : 0;
            }}
          />
          <Table.Column
            dataIndex="createTime"
            title={i18n.t('Create Time').toString()}
            cell={(v: string) => {
              return momentDate(v);
            }}
          />
          <Table.Column
            dataIndex="updateTime"
            title={i18n.t('Update Time').toString()}
            cell={(v: string) => {
              return momentDate(v);
            }}
          />
          <Table.Column
            dataIndex="name"
            title={i18n.t('Action').toString()}
            width={200}
            lock="right"
            cell={(v: string, i: number, w: Workflow) => {
              // The workflow an environment was given on binding goes with it.
              const envDefault = v == 'workflow-' + w.envName;
              return (
                <RowLinks
                  links={[
                    {
                      label: 'Edit',
                      hidden: applicationDetail?.readOnly || !may(w, 'update'),
                      onClick: () => this.props.dispatch(routerRedux.push(studio(w))),
                    },
                    {
                      label: 'Remove',
                      danger: true,
                      hidden: applicationDetail?.readOnly || !may(w, 'delete'),
                      disabled: envDefault,
                      reason: "An environment's own workflow goes with the environment",
                      onClick: () => this.onDeleteWorkflow(v),
                    },
                  ]}
                />
              );
            }}
          />
        </Table>
        <TablePager
          total={list.length}
          page={page}
          pageSize={pageSize}
          onChange={(p, size) => this.setState({ page: p, pageSize: size })}
        />
      </div>
    );
  }
}

export default ApplicationWorkflowList;
