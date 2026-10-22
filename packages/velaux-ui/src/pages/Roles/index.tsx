import { Button, Dialog, Message, Table } from '@alifd/next';
import { connect } from 'dva';
import React, { Component, Fragment } from 'react';

import { getPlatformPermissions } from '../../api/rbac';
import { deleteRole, getRoleList } from '../../api/roles';
import { If } from '../../components/If';
import { ListTitle as Title } from '../../components/ListTitle';
import { ListToolbar, RowLinks, TablePager } from '../../components/Console';
import Permission from '../../components/Permission';
import { Translation } from '../../components';
import type { LoginUserInfo, PermissionBase, RolesBase } from '@velaux/data';
import { momentDate } from '../../utils/common';
import { locale } from '../../utils/locale';
import { checkPermission } from '../../utils/permission';
import './index.less';

import RolesDialog from './components/RolesDialog';

interface Props {
  userInfo?: LoginUserInfo;
}

type State = {
  list: RolesBase[];
  total: number;
  page: number;
  pageSize: number;
  permissions: PermissionBase[];
  visible: boolean;
  isEditRole: boolean;
  editRoleItem: RolesBase;
  isLoading: boolean;
  // query narrows the loaded page by name or alias; the list call takes none.
  query: string;
};

@connect((store: any) => {
  return { ...store.user };
})
class Roles extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      list: [],
      total: 0,
      page: 1,
      pageSize: 10,
      permissions: [],
      visible: false,
      isEditRole: false,
      editRoleItem: { name: '' },
      isLoading: false,
      query: '',
    };
  }

  componentDidMount() {
    this.listRoles();
    this.listPermissions();
  }

  listRoles = async () => {
    this.setState({ isLoading: true });
    const { page, pageSize } = this.state;
    getRoleList({ page, pageSize })
      .then((res) => {
        this.setState({
          list: (res && res.roles) || [],
          total: (res && res.total) || 0,
          isLoading: false,
        });
      })
      .catch(() => {
        this.setState({ isLoading: false });
      });
  };

  listPermissions = async () => {
    getPlatformPermissions().then((res) => {
      this.setState({
        permissions: (res && res) || [],
      });
    });
  };

  onEdit = (record: RolesBase) => {
    this.setState({
      visible: true,
      isEditRole: true,
      editRoleItem: record,
    });
  };

  onDelete = (record: RolesBase) => {
    Dialog.confirm({
      content: 'Are you sure you want to delete this role',
      onOk: () => {
        const { name } = record;
        if (name) {
          deleteRole({ name }).then((res) => {
            if (res) {
              Message.success(<Translation>Delete the role success</Translation>);
              this.listRoles();
            }
          });
        }
      },
      locale: locale().Dialog,
    });
  };

  onCreate = () => {
    this.setState({ visible: false });
    this.listRoles();
  };

  onCloseCreate = () => {
    this.setState({ visible: false });
  };

  handleClickCreate = () => {
    this.setState({
      visible: true,
      isEditRole: false,
    });
  };

  handleChange = (page: number, pageSize: number) => {
    this.setState(
      {
        page,
        pageSize,
      },
      () => {
        this.listRoles();
      }
    );
  };

  render() {
    const columns = [
      {
        key: 'name',
        title: <Translation>Name</Translation>,
        dataIndex: 'name',
        cell: (v: string) => {
          return <span>{v}</span>;
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
        key: 'createTime',
        title: <Translation>Create Time</Translation>,
        dataIndex: 'createTime',
        cell: (v: string) => {
          return <span>{momentDate(v)}</span>;
        },
      },
      // {
      //   key: 'permPolicies',
      //   title: <Translation>PermPolicies</Translation>,
      //   dataIndex: 'permPolicies',
      //   cell: (v: any) => {
      //     return v.map((item: any) => (<span className='roles-permPolicies margin-right-5'>{item && item.name}</span>))
      //   },
      // },
      {
        key: 'operation',
        title: <Translation>Actions</Translation>,
        dataIndex: 'operation',
        width: 160,
        lock: 'right',
        cell: (v: string, i: number, record: RolesBase) => {
          const may = (action: string) =>
            checkPermission({ resource: `role:${record.name}`, action }, '', this.props.userInfo);
          return (
            <RowLinks
              links={[
                { label: 'Edit', hidden: !may('update'), onClick: () => this.onEdit(record) },
                { label: 'Delete', danger: true, hidden: !may('delete'), onClick: () => this.onDelete(record) },
              ]}
            />
          );
        },
      },
    ];

    const { Column } = Table;
    const { list, page, pageSize, total, visible, isEditRole, editRoleItem, permissions, isLoading, query } =
      this.state;
    const keyword = query.trim().toLowerCase();
    const shown = keyword
      ? list.filter((r) => r.name.toLowerCase().includes(keyword) || (r.alias || '').toLowerCase().includes(keyword))
      : list;
    return (
      <Fragment>
        <div className="roles-list-content">
          <Title
            title="Platform Roles"
            subTitle="Assign permissions for resources such as clusters、targets、addons、projects and users"
          />
          <div className="console-panel">
            <div className="console-panel-body">
              <ListToolbar
                search={{
                  value: query,
                  placeholder: 'Search by name or alias',
                  live: true,
                  onSearch: (value: string) => this.setState({ query: value }),
                }}
                onRefresh={this.listRoles}
                refreshing={isLoading}
                actions={
                  <Permission request={{ resource: 'role:*', action: 'create' }} project={''}>
                    <Button type="primary" onClick={this.handleClickCreate}>
                      <Translation>New Role</Translation>
                    </Button>
                  </Permission>
                }
              />
            </div>
            <Table locale={locale().Table} dataSource={shown} loading={isLoading} primaryKey="name">
              {columns.map((col, key) => (
                <Column {...col} key={key} align={'left'} />
              ))}
            </Table>
            <TablePager total={total} page={page} pageSize={pageSize} onChange={this.handleChange} />
          </div>

          <If condition={visible}>
            <RolesDialog
              visible={visible}
              isEditRole={isEditRole}
              editRoleItem={editRoleItem}
              permissions={permissions}
              onCreate={this.onCreate}
              onCloseCreate={this.onCloseCreate}
            />
          </If>
        </div>
      </Fragment>
    );
  }
}

export default Roles;
