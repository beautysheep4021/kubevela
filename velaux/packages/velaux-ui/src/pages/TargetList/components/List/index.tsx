import { Table, Message, Dialog } from '@alifd/next';
import React, { Component } from 'react';
import { connect } from 'dva';

import { deleteTarget } from '../../../../api/target';
import { RowLinks } from '../../../../components/Console';
import { Translation } from '../../../../components/Translation';
import './index.less';
import type { LoginUserInfo, Project, Target } from '@velaux/data';
import { locale } from '../../../../utils/locale';
import { checkPermission } from '../../../../utils/permission';

import { Link } from 'dva/router';

type Props = {
  list?: Target[];
  userInfo?: LoginUserInfo;
  updateTargetList: () => void;
  changeISEdit: (param: boolean, record: Target) => void;
};

@connect((store: any) => {
  return { ...store.user };
})
class TableList extends Component<Props> {
  onDelete = (record: Target) => {
    deleteTarget({ name: record.name || '' }).then((re) => {
      if (re) {
        Message.success('target delete success');
        this.props.updateTargetList();
      }
    });
  };

  onEdit = (record: Target) => {
    this.props.changeISEdit(true, record);
  };

  getColumns = () => {
    return [
      {
        key: 'name',
        title: <Translation>Name</Translation>,
        dataIndex: 'name',
        cell: (v: string) => {
          return <span className="target-table-name">{v}</span>;
        },
      },
      {
        key: 'alias',
        title: <Translation>Alias</Translation>,
        dataIndex: 'alias',
        cell: (v: string) => {
          return <span>{v}</span>;
        },
      },
      {
        key: 'project',
        title: <Translation>Project</Translation>,
        dataIndex: 'project',
        cell: (v: Project) => {
          if (v && v.name) {
            return <Link to={`/projects/${v.name}/summary`}>{v && v.name}</Link>;
          } else {
            return null;
          }
        },
      },
      {
        key: 'clusterName/namespace',
        title: <Translation>Cluster/Namespace</Translation>,
        dataIndex: 'clusterName/CloudProvider',
        cell: (v: string, i: number, record: Target) => {
          return (
            <span>
              {record?.clusterAlias ? record?.clusterAlias : record?.cluster?.clusterName}/{record?.cluster?.namespace}
            </span>
          );
        },
      },
      {
        key: 'description',
        title: <Translation>Description</Translation>,
        dataIndex: 'description',
        cell: (v: string) => {
          return <span>{v}</span>;
        },
      },
      {
        key: 'operation',
        title: <Translation>Actions</Translation>,
        dataIndex: 'operation',
        width: 160,
        lock: 'right',
        cell: (v: string, i: number, record: Target) => {
          const may = (action: string) =>
            checkPermission({ resource: `target:${record.name}`, action }, '', this.props.userInfo);
          return (
            <RowLinks
              links={[
                { label: 'Edit', hidden: !may('update'), onClick: () => this.onEdit(record) },
                {
                  label: 'Remove',
                  danger: true,
                  hidden: !may('delete'),
                  onClick: () =>
                    Dialog.confirm({
                      type: 'confirm',
                      content: <Translation>Unrecoverable after deletion, are you sure to delete it?</Translation>,
                      onOk: () => this.onDelete(record),
                      locale: locale().Dialog,
                    }),
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
    return (
      <div className="table-delivery-list">
        <Table locale={locale().Table} size="medium" dataSource={list} primaryKey="name" loading={false}>
          {columns.map((col) => (
            <Column {...col} key={col.key} align={'left'} />
          ))}
        </Table>
      </div>
    );
  }
}

export default TableList;
