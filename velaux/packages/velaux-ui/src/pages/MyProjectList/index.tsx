import { Table, Dialog, Message, Tag } from '@alifd/next';
import { Link } from 'dva/router';
import React, { Fragment, Component } from 'react';

import { deleteProject } from '../../api/project';
import { ListToolbar, RowLinks, TablePager, pageOf } from '../../components/Console';
import { ListTitle as Title } from '../../components/ListTitle';
import { Translation } from '../../components/Translation';
import type { NameAlias, Project } from '@velaux/data';
import { momentDate } from '../../utils/common';
import { locale } from '../../utils/locale';
import { connect } from 'dva';
import { LoginUserInfo } from '@velaux/data';
import i18n from '../../i18n';
import { locationService } from '../../services/LocationService';

type Props = {
  dispatch: ({}) => void;
  userInfo?: LoginUserInfo;
};

type State = {
  isLoading: boolean;
  query: string;
  page: number;
  pageSize: number;
};

@connect((store: any) => {
  return { ...store.user };
})
class MyProjectList extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      isLoading: false,
      query: '',
      page: 1,
      pageSize: 10,
    };
  }
  componentDidMount() {}

  loadUserInfo = () => {
    this.setState({ isLoading: true });
    this.props.dispatch({
      type: 'user/getLoginUserInfo',
      callback: () => {
        this.setState({ isLoading: false });
      },
    });
  };

  onDelete = (record: Project) => {
    Dialog.confirm({
      content: 'Are you sure you want to delete the project',
      onOk: () => {
        const { name } = record;
        if (name) {
          deleteProject({ name })
            .then((res) => {
              if (res) {
                Message.success(<Translation>Project deleted successfully</Translation>);
                this.loadUserInfo();
              }
            })
            .catch();
        }
      },
      locale: locale().Dialog,
    });
  };

  render() {
    const columns = [
      {
        key: 'name',
        title: <Translation>Name</Translation>,
        dataIndex: 'name',
        cell: (v: string) => {
          return (
            <span>
              <Link to={`/projects/${v}`}>{v}</Link>
            </span>
          );
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
        key: 'description',
        title: <Translation>Description</Translation>,
        dataIndex: 'description',
        cell: (v: string) => {
          return <span>{v}</span>;
        },
      },
      {
        key: 'owner',
        title: <Translation>Owner</Translation>,
        dataIndex: 'owner',
        cell: (v: NameAlias) => {
          return <span>{v && v.alias ? `${v.alias}(${v.name})` : v.name}</span>;
        },
      },
      {
        key: 'joinTime',
        title: <Translation>Join Time</Translation>,
        dataIndex: 'joinTime',
        cell: (v: string) => {
          return <span>{momentDate(v)}</span>;
        },
      },
      {
        key: 'roles',
        title: <Translation>Roles</Translation>,
        dataIndex: 'roles',
        cell: (v: NameAlias[]) => {
          if (!v) {
            return <span></span>;
          }
          return (
            <div>
              {v.map((role) => {
                return <Tag key={role.name}>{role.alias || role.name}</Tag>;
              })}
            </div>
          );
        },
      },
      {
        key: 'operation',
        title: <Translation>Actions</Translation>,
        dataIndex: 'operation',
        width: 180,
        lock: 'right',
        cell: (v: string, i: number, record: Project) => (
          <RowLinks
            links={[
              { label: 'Details', onClick: () => locationService.push(`/projects/${record.name}`) },
              { label: 'Applications', onClick: () => locationService.push(`/projects/${record.name}/applications`) },
            ]}
          />
        ),
      },
    ];

    const { Column } = Table;
    const { isLoading, query, pageSize } = this.state;
    const { userInfo } = this.props;
    // The search narrows the projects the user is in, by name or alias.
    const keyword = query.trim().toLowerCase();
    const projects = (userInfo?.projects || []).filter(
      (p) => !keyword || p.name.toLowerCase().includes(keyword) || (p.alias || '').toLowerCase().includes(keyword)
    );
    const page = Math.min(this.state.page, Math.max(1, Math.ceil(projects.length / pageSize)));
    return (
      <Fragment>
        <div className="project-list-content">
          <Title title="Projects" subTitle={i18n.t('Project are used for users and applications isolation')} />
          <div className="console-panel">
            <div className="console-panel-body">
              <ListToolbar
                search={{
                  value: query,
                  placeholder: 'Search by name or alias',
                  live: true,
                  onSearch: (value: string) => this.setState({ query: value, page: 1 }),
                }}
                onRefresh={this.loadUserInfo}
                refreshing={isLoading}
              />
            </div>
            <Table
              locale={locale().Table}
              dataSource={pageOf(projects, page, pageSize)}
              loading={isLoading}
              primaryKey="name"
            >
              {columns.map((col, key) => (
                <Column {...col} key={key} align={'left'} />
              ))}
            </Table>
            <TablePager
              total={projects.length}
              page={page}
              pageSize={pageSize}
              onChange={(p, size) => this.setState({ page: p, pageSize: size })}
            />
          </div>
        </div>
      </Fragment>
    );
  }
}

export default MyProjectList;
