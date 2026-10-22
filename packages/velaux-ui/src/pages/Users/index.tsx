import { Table, Button, Message, Dialog, Input } from '@alifd/next';
import { connect } from 'dva';
import React, { Component, Fragment } from 'react';

import { getLoginType } from '../../api/authentication';
import { getRoleList } from '../../api/roles';
import { getUserList, deleteUser, changeUserDisable, changeUserEnable } from '../../api/users';
import { If } from '../../components/If';
import { ListTitle } from '../../components/ListTitle';
import { ListToolbar, RowLinks, TablePager } from '../../components/Console';
import Permission from '../../components/Permission';
import { StatusBadge } from '../../components/StatusBadge';
import { Translation } from '../../components/Translation';
import type { LoginUserInfo, NameAlias, RolesBase, User } from '@velaux/data';
import i18n from '../../i18n';
import { momentDate } from '../../utils/common';
import { locale } from '../../utils/locale';
import { checkPermission } from '../../utils/permission';

import CreateUser from './components/CreateUser';
import ResetPassword from './components/ResetPassword';

import './index.less';

type Props = {
  userInfo?: LoginUserInfo;
};

type State = {
  name: string;
  email: string;
  alias: string;
  page: number;
  pageSize: number;
  isLoading: boolean;
  dataSource: [];
  total: number;
  isUserDialogVisible: boolean;
  isEditUser: boolean;
  editUser: User;
  rolesList: RolesBase[];
  isResetPassword: boolean;
  isResetPasswordDialog: boolean;
  // dex is whether users sign in through Dex, which owns their passwords.
  dex: boolean;
};

@connect((store: any) => {
  return { ...store.user };
})
class Users extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      isUserDialogVisible: false,
      dex: false,
      name: '',
      email: '',
      alias: '',
      page: 1,
      pageSize: 10,
      isLoading: false,
      total: 0,
      dataSource: [],
      rolesList: [],
      isEditUser: false,
      isResetPassword: false,
      isResetPasswordDialog: false,
      editUser: {
        name: '',
        email: '',
        disabled: false,
      },
    };
  }

  componentDidMount() {
    this.listUser();
    this.listRoles();
    // Unknown, the action stays offered; in Dex mode the server refuses it.
    getLoginType()
      .then((res: { loginType?: string }) => {
        this.setState({ dex: res?.loginType === 'dex' });
      })
      .catch(() => {});
  }

  listUser = async () => {
    const { name, alias, email, page, pageSize } = this.state;
    const params = {
      name,
      email,
      alias,
      page,
      pageSize,
    };
    this.setState({ isLoading: true });
    getUserList(params)
      .then((res) => {
        if (res) {
          this.setState({
            dataSource: res.users || [],
            total: res.total,
          });
        }
      })
      .finally(() => {
        this.setState({
          isLoading: false,
        });
      });
  };

  listRoles = async () => {
    getRoleList({}).then((res) => {
      this.setState({
        rolesList: (res && res.roles) || [],
      });
    });
  };

  handleChangName = (e: string) => {
    this.setState(
      {
        name: e,
      },
      () => {
        this.listUser();
      }
    );
  };

  handleChangAlias(e: string) {
    this.setState(
      {
        alias: e,
      },
      () => {
        this.listUser();
      }
    );
  }

  handleChangEmail(e: string) {
    this.setState(
      {
        email: e,
      },
      () => {
        this.listUser();
      }
    );
  }

  onEdit = (record: User) => {
    this.onEditUser(record);
  };

  onResetPassword = (record: User) => {
    this.onEditResetPasswordUser(record);
  };

  onCreate = () => {
    this.setState({
      isUserDialogVisible: false,
    });
    this.listUser();
  };

  onResetPassWord = () => {
    this.setState({
      isResetPasswordDialog: false,
    });
    this.listUser();
  };

  onChangeStatus = (record: User) => {
    const { disabled, name } = record;
    if (disabled) {
      changeUserEnable({ name })
        .then((res) => {
          if (res) {
            Message.success(<Translation>Update user status success</Translation>);
            this.listUser();
          }
        })
        .catch();
    } else {
      changeUserDisable({ name })
        .then((res) => {
          if (res) {
            Message.success(<Translation>Update user status success</Translation>);
            this.listUser();
          }
        })
        .catch();
    }
  };

  onDelete = (record: User) => {
    Dialog.confirm({
      content: 'Are you sure you want to delete the user',
      onOk: () => {
        const { name } = record;
        if (name) {
          deleteUser({ name })
            .then((res) => {
              if (res) {
                Message.success(<Translation>Delete user success</Translation>);
                this.listUser();
              }
            })
            .catch();
        }
      },
      locale: locale().Dialog,
    });
  };

  onClose = () => {
    this.setState({
      isUserDialogVisible: false,
    });
  };

  onResetPassWordClose = () => {
    this.setState({
      isResetPasswordDialog: false,
    });
  };

  onEditUser = (editUser: User) => {
    this.setState({
      editUser: editUser,
      isEditUser: true,
      isUserDialogVisible: true,
    });
  };

  onEditResetPasswordUser = (editUser: User) => {
    this.setState({
      editUser: editUser,
      isResetPasswordDialog: true,
      isResetPassword: true,
    });
  };

  handleClickCreate = () => {
    this.setState({
      isUserDialogVisible: true,
      isEditUser: false,
    });
  };

  handleChange = (page: number, pageSize: number) => {
    this.setState(
      {
        page,
        pageSize,
      },
      () => {
        this.listUser();
      }
    );
  };

  render() {
    const { Column } = Table;
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
        key: 'roles',
        title: <Translation>Platform Roles</Translation>,
        dataIndex: 'roles',
        cell: (v: NameAlias[]) => {
          return (v || []).map((item: NameAlias) => (
            <span className="roles-permPolicies margin-right-5">{item.alias || item.name}</span>
          ));
        },
      },
      {
        key: 'status',
        title: <Translation>Status</Translation>,
        dataIndex: 'disabled',
        cell: (disabled: boolean) => (
          <StatusBadge tone={disabled ? 'neutral' : 'healthy'} label={disabled ? 'Disabled' : 'Active'} />
        ),
      },
      {
        key: 'email',
        title: <Translation>Email</Translation>,
        dataIndex: 'email',
        cell: (v: string) => {
          return <span>{v}</span>;
        },
      },
      {
        key: 'createTime',
        title: <Translation>CreateTime</Translation>,
        dataIndex: 'createTime',
        cell: (v: string) => {
          if (v) {
            return <span>{momentDate(v)}</span>;
          }
          return '';
        },
      },
      {
        key: 'lastLoginTime',
        title: <Translation>LastLoginTime</Translation>,
        dataIndex: 'lastLoginTime',
        cell: (v: string) => {
          if (v) {
            return <span>{momentDate(v)}</span>;
          }
          return '';
        },
      },
      {
        key: 'operation',
        title: <Translation>Actions</Translation>,
        dataIndex: 'operation',
        width: 260,
        lock: 'right',
        cell: (v: string, i: number, record: User) => {
          const may = (action: string) =>
            checkPermission({ resource: `user:${record.name}`, action }, '', this.props.userInfo);
          return (
            <RowLinks
              links={[
                { label: 'Edit', hidden: !may('update'), onClick: () => this.onEdit(record) },
                {
                  label: 'Reset Password',
                  // Dex owns the passwords of users who sign in through it.
                  hidden: this.state.dex || !may('update'),
                  onClick: () => this.onResetPassword(record),
                },
                {
                  label: record.disabled ? 'Enable' : 'Disable',
                  danger: !record.disabled,
                  hidden: !may(record.disabled ? 'enable' : 'disable'),
                  onClick: () => this.onChangeStatus(record),
                },
                { label: 'Delete', danger: true, hidden: !may('delete'), onClick: () => this.onDelete(record) },
              ]}
            />
          );
        },
      },
    ];

    const {
      name,
      alias,
      email,
      dataSource,
      isUserDialogVisible,
      isEditUser,
      editUser,
      page,
      pageSize,
      total,
      isLoading,
      isResetPasswordDialog,
      isResetPassword,
      rolesList,
    } = this.state;
    return (
      <Fragment>
        <div className="user-wrapper">
          <ListTitle
            title="Users"
            subTitle="Basic authorization management is provided for local users by default, but SSO authentication is strongly recommended"
          />

          <div className="console-panel user-list-wrapper">
            <div className="console-panel-body">
              <ListToolbar
                search={{
                  value: name,
                  placeholder: 'Search by name',
                  live: true,
                  onSearch: (nameValue: string) => this.handleChangName(nameValue),
                }}
                filters={
                  <Fragment>
                    <Input
                      hasClear
                      className="user-filter"
                      placeholder={i18n.t('Search by alias').toString()}
                      onChange={(aliasValue: string) => this.handleChangAlias(aliasValue)}
                      value={alias}
                    />
                    <Input
                      hasClear
                      className="user-filter"
                      placeholder={i18n.t('Search by email').toString()}
                      onChange={(emailValue: string) => this.handleChangEmail(emailValue)}
                      value={email}
                    />
                  </Fragment>
                }
                onRefresh={this.listUser}
                refreshing={isLoading}
                actions={
                  <Permission request={{ resource: 'user:*', action: 'create' }} project={''}>
                    <Button type="primary" onClick={this.handleClickCreate}>
                      <Translation>New User</Translation>
                    </Button>
                  </Permission>
                }
              />
            </div>
            <Table locale={locale().Table} dataSource={dataSource} loading={isLoading} primaryKey="name">
              {columns.map((col, key) => (
                <Column {...col} key={key} align={'left'} />
              ))}
            </Table>
            <TablePager total={total} page={page} pageSize={pageSize} onChange={this.handleChange} />
          </div>

          <If condition={isUserDialogVisible}>
            <CreateUser
              visible={isUserDialogVisible}
              isEditUser={isEditUser}
              editUser={editUser}
              rolesList={rolesList}
              onClose={this.onClose}
              onCreate={this.onCreate}
            />
          </If>

          <If condition={isResetPasswordDialog}>
            <ResetPassword
              visible={isResetPasswordDialog}
              editUser={editUser}
              isResetPassword={isResetPassword}
              onClose={this.onResetPassWordClose}
              onCreate={this.onResetPassWord}
            />
          </If>
        </div>
      </Fragment>
    );
  }
}

export default Users;
