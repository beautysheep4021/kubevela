import React from 'react';
import './index.less';

import { Dialog, Table } from '@alifd/next';

import kubernetesSvg from '../../../../assets/kubernetes.svg';
import { RowLinks, TablePager, pageOf } from '../../../../components/Console';
import { StatusBadge } from '../../../../components/StatusBadge';
import { Translation } from '../../../../components/Translation';
import type { Cluster, LoginUserInfo } from '@velaux/data';
import { locale } from '../../../../utils/locale';
import { momentDate } from '../../../../utils/common';
import { checkPermission } from '../../../../utils/permission';

import { connect } from 'dva';

type State = {
  page: number;
  pageSize: number;
};

type Props = {
  clusters: [];
  userInfo?: LoginUserInfo;
  editCluster: (name: string) => void;
  deleteCluster: (name: string) => void;
};

// The control plane's own cluster, which can be neither edited nor detached.
const localCluster = 'local';

// CardContent lists the clusters as a table, a page at a time.
@connect((store: any) => {
  return { ...store.user };
})
class CardContent extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      page: 1,
      pageSize: 10,
    };
  }
  editCluster = (name: string) => {
    this.props.editCluster(name);
  };
  onDeleteCluster = (name: string) => {
    this.props.deleteCluster(name);
  };

  // rowLinks are a cluster's operations: its dashboard, and the edits the user
  // may make; the local cluster offers them disabled.
  rowLinks = (item: Cluster) => {
    const { userInfo } = this.props;
    const may = (action: string) => checkPermission({ resource: `cluster:${item.name}`, action }, '', userInfo);
    const local = item.name === localCluster;
    return (
      <RowLinks
        links={[
          {
            label: 'Dashboard',
            disabled: !item.dashboardURL,
            reason: 'No dashboard URL is set for this cluster',
            onClick: () => window.open(item.dashboardURL, '_blank', 'noopener,noreferrer'),
          },
          {
            label: 'Edit',
            hidden: !may('update'),
            disabled: local,
            reason: 'The local cluster cannot be changed',
            onClick: () => this.editCluster(item.name),
          },
          {
            label: 'Detach',
            danger: true,
            hidden: !may('delete'),
            disabled: local,
            reason: 'The local cluster cannot be changed',
            onClick: () =>
              Dialog.confirm({
                type: 'confirm',
                content: <Translation>Are you sure you want the detach cluster?</Translation>,
                onOk: () => this.onDeleteCluster(item.name),
                locale: locale().Dialog,
              }),
          },
        ]}
      />
    );
  };

  getColumns = () => {
    return [
      {
        key: 'name',
        title: <Translation>Name</Translation>,
        dataIndex: 'name',
        cell: (v: string, i: number, item: Cluster) => (
          <span className="cluster-table-name">
            <img src={item.icon && item.icon !== 'none' ? item.icon : kubernetesSvg} />
            <span>
              {item.dashboardURL ? (
                <a target="_blank" href={item.dashboardURL} rel="noopener noreferrer">
                  {v}
                </a>
              ) : (
                <span className="cluster-table-title">{v}</span>
              )}
              {item.alias && item.alias !== v && <span className="cluster-table-alias">{item.alias}</span>}
            </span>
          </span>
        ),
      },
      {
        key: 'status',
        title: <Translation>Status</Translation>,
        dataIndex: 'status',
        cell: (v: string) => (
          <StatusBadge
            tone={v === 'Healthy' ? 'healthy' : v === 'Unhealthy' ? 'unhealthy' : 'neutral'}
            label={v || 'Unknown'}
          />
        ),
      },
      {
        key: 'provider',
        title: <Translation>Provider</Translation>,
        dataIndex: 'providerInfo',
        cell: (v: Cluster['providerInfo'], i: number, item: Cluster) => {
          if (item.name === localCluster) {
            return <Translation>Local</Translation>;
          }
          if (!v?.provider) {
            return '-';
          }
          return (
            <span>
              {v.provider}
              {v.regionID && <span className="cluster-table-alias">{v.regionID}</span>}
            </span>
          );
        },
      },
      {
        key: 'labels',
        title: <Translation>Labels</Translation>,
        dataIndex: 'labels',
        cell: (v?: Record<string, string>) => {
          const keys = Object.keys(v || {});
          if (keys.length === 0) {
            return '-';
          }
          return (
            <span className="cluster-table-labels">
              {keys.map((key) => (
                <span key={key} className="cluster-table-label">{`${key}=${v?.[key]}`}</span>
              ))}
            </span>
          );
        },
      },
      {
        key: 'description',
        title: <Translation>Description</Translation>,
        dataIndex: 'description',
        cell: (v: string) => <span className="cluster-table-description">{v || '-'}</span>,
      },
      {
        key: 'createTime',
        title: <Translation>Created</Translation>,
        dataIndex: 'createTime',
        cell: (v: string) => <span className="cluster-table-nowrap">{v ? momentDate(v) : '-'}</span>,
      },
      {
        key: 'operation',
        title: <Translation>Actions</Translation>,
        dataIndex: 'operation',
        width: 200,
        lock: 'right',
        cell: (v: string, i: number, item: Cluster) => this.rowLinks(item),
      },
    ];
  };

  render() {
    const { Column } = Table;
    const { clusters } = this.props;
    const { pageSize } = this.state;
    // A search that shrinks the list keeps the page in range.
    const page = Math.min(this.state.page, Math.max(1, Math.ceil(clusters.length / pageSize)));
    return (
      <div className="cluster-table">
        <Table
          locale={locale().Table}
          size="medium"
          dataSource={pageOf(clusters, page, pageSize)}
          primaryKey="name"
          loading={false}
        >
          {this.getColumns().map((col) => (
            <Column {...col} key={col.key} align={'left'} />
          ))}
        </Table>
        <TablePager
          total={clusters.length}
          page={page}
          pageSize={pageSize}
          onChange={(p, size) => this.setState({ page: p, pageSize: size })}
        />
      </div>
    );
  }
}

export default CardContent;
