import { ColumnProps } from '@alifd/next/types/table';
import { Table, Message, Dialog, Tag } from '@alifd/next';
import { Link } from 'dva/router';
import React, { Component } from 'react';

import { deleteEnv } from '../../../../api/env';
import { If } from '../../../../components/If';
import { RowLinks } from '../../../../components/Console';
import { Translation } from '../../../../components/Translation';
import './index.less';
import type { Env, NameAlias, Project, LoginUserInfo } from '@velaux/data';
import { locale } from '../../../../utils/locale';
import { checkPermission } from '../../../../utils/permission';
const { Group: TagGroup } = Tag;

type Props = {
  list?: Env[];
  updateEnvList: () => void;
  userInfo?: LoginUserInfo;
  changeISEdit: (param: boolean, record: Env) => void;
};

class TableList extends Component<Props> {
  onDelete = (record: Env) => {
    deleteEnv({ name: record.name || '' }).then((re) => {
      if (re) {
        Message.success('Environment deleted successfully');
        this.props.updateEnvList();
      }
    });
  };
  onEdit = (record: Env) => {
    this.props.changeISEdit(true, record);
  };

  getColumns = (): Array<ColumnProps & { key: string; lock?: string }> => {
    return [
      {
        key: 'name',
        title: <Translation>Name(Alias)</Translation>,
        dataIndex: 'name',
        width: '200px',
        cell: (v: string, i: number, env: Env) => {
          return (
            <span>
              <span className="env-table-name">{env.name}</span>
              {env.alias && env.alias !== env.name && <span className="env-table-alias">{env.alias}</span>}
            </span>
          );
        },
      },
      {
        key: 'project',
        title: <Translation>Project</Translation>,
        dataIndex: 'project',
        width: '100px',
        cell: (v: Project) => {
          if (v && v.name) {
            return <Link to={`/projects/${v.name}/summary`}>{v.alias || v.name}</Link>;
          } else {
            return null;
          }
        },
      },
      {
        key: 'namespace',
        title: <Translation>Namespace</Translation>,
        dataIndex: 'namespace',
        width: '120px',
        cell: (v: string) => {
          return <span>{v}</span>;
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
        key: 'targets',
        title: <Translation>Targets</Translation>,
        dataIndex: 'targets',
        cell: (v: NameAlias[], i: number, record: Env) => {
          return (
            <TagGroup>
              {v?.map((target: NameAlias) => {
                return (
                  <Tag color="green" key={target.name + i}>
                    <If
                      condition={checkPermission(
                        { resource: `target:*`, action: 'list' },
                        record.project.name,
                        this.props.userInfo
                      )}
                    >
                      <Link to="/targets">{target.alias ? target.alias : target.name}</Link>
                    </If>
                    <If
                      condition={
                        !checkPermission(
                          { resource: `target:*`, action: 'list' },
                          record.project.name,
                          this.props.userInfo
                        )
                      }
                    >
                      {target.alias ? target.alias : target.name}
                    </If>
                  </Tag>
                );
              })}
            </TagGroup>
          );
        },
      },
      {
        key: 'operation',
        title: <Translation>Actions</Translation>,
        dataIndex: 'operation',
        width: 160,
        lock: 'right',
        cell: (v: string, i: number, record: Env) => {
          const project = record.project.name;
          const may = (action: string) =>
            checkPermission(
              { resource: `project:${project}/environment:${record.name}`, action },
              project,
              this.props.userInfo
            );
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
      <div className="table-env-list">
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
