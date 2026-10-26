import { Table, Message, Tag, Balloon } from '@alifd/next';
import i18n from 'i18next';
import { connect } from 'dva';
import { Link } from 'dva/router';
import _ from 'lodash';
import React, { Component } from 'react';

import { getDefinitionsList, updateDefinitionStatus } from '../../api/definitions';
import { ListToolbar, RowLinks, TablePager, pageOf } from '../../components/Console';
import { RestrictionTags } from '../../components/RestrictionTags';
import { StatusBadge } from '../../components/StatusBadge';
import { Translation } from '../../components/Translation';
import type { DefinitionBase, DefinitionRestrictions, LoginUserInfo } from '@velaux/data';

// import { momentDate } from '../../utils/common';

import { locale } from '../../utils/locale';
import { getMatchParamObj } from '../../utils/utils';

import { PolicyScopeTag } from '../../components/PolicyScopeTag';
import { UsageDialog } from './components/UsageDialog';

import './index.less';
import { checkPermission } from '../../utils/permission';
import { projectChanged } from '../../utils/currentProject';
import type { DefinitionPlace } from '../../utils/definitionPlace';
import { definitionPlaceQuery, definitionResource } from '../../utils/definitionPlace';

type Props = {
  match: {
    params: {
      definitionType: 'component' | 'trait' | 'workflowstep' | 'policy';
    };
  };
  userInfo?: LoginUserInfo;
  currentProject?: { current: string; resolved: boolean };
};

type State = {
  definitionType: 'component' | 'trait' | 'workflowstep' | 'policy';
  list: DefinitionBase[];
  isLoading: boolean;
  searchValue: string;
  searchList: DefinitionBase[];
  page: number;
  pageSize: number;
  // usageOf is the definition whose quota usage is shown.
  usageOf?: DefinitionBase;
};

@connect((store: any) => {
  return { ...store.definitions, ...store.user, currentProject: store.currentProject };
})
class Definitions extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      definitionType: this.getDefinitionType(),
      list: [],
      isLoading: false,
      searchValue: '',
      searchList: [],
      page: 1,
      pageSize: 20,
    };
  }
  componentDidMount() {
    this.lisDefinitions();
  }

  componentWillReceiveProps(nextProps: Props) {
    // The list waits on the user and their permissions, which load after the page.
    if (
      nextProps.userInfo !== this.props.userInfo ||
      projectChanged(this.props.currentProject, nextProps.currentProject)
    ) {
      this.lisDefinitions(nextProps.userInfo, nextProps.currentProject);
    }
    const nextPropsParams = nextProps.match.params || {};
    if (nextPropsParams.definitionType !== this.state.definitionType) {
      this.setState(
        {
          definitionType: nextPropsParams.definitionType,
        },
        () => {
          this.lisDefinitions();
        }
      );
    }
  }

  // project is the one picked in the top bar: its own definitions are listed
  // with the global ones. Every project at once lists the global ones alone.
  project = (currentProject = this.props.currentProject) => (currentProject?.resolved ? currentProject.current : '');

  // place is where a listed definition is.
  place = (record: DefinitionBase): DefinitionPlace => ({
    project: this.project(),
    where: record.scope === 'project' ? 'project' : 'global',
  });

  lisDefinitions(userInfo = this.props.userInfo, currentProject = this.props.currentProject) {
    const { definitionType } = this.state;
    if (!definitionType) {
      return;
    }
    const project = this.project(currentProject);
    const resource = project ? `project:${project}/definition:*` : 'definition:*';
    if (!checkPermission({ resource, action: 'list' }, project, userInfo)) {
      return;
    }
    const params = {
      project,
      definitionType,
      queryAll: true,
    };
    this.setState({ isLoading: true });
    getDefinitionsList(params)
      .then((res) => {
        if (res) {
          this.setState({
            list: (res && res.definitions) || [],
            searchList: [],
            searchValue: '',
          });
        } else {
          this.setState({
            list: [],
            searchList: [],
            searchValue: '',
          });
        }
      })
      .finally(() => {
        this.setState({ isLoading: false });
      });
  }

  getDefinitionType = () => {
    return getMatchParamObj(this.props.match, 'definitionType');
  };

  onChangeStatus = (record: DefinitionBase) => {
    const { definitionType } = this.state;
    const { status, name } = record;
    if (status === 'enable') {
      updateDefinitionStatus({ ...this.place(record), name, hiddenInUI: true, type: definitionType })
        .then((res) => {
          if (res) {
            Message.success(<Translation>Update definition status success</Translation>);
            this.lisDefinitions();
          }
        })
        .catch();
    } else {
      updateDefinitionStatus({ ...this.place(record), name, hiddenInUI: false, type: definitionType })
        .then((res) => {
          if (res) {
            Message.success(<Translation>Update definition status success</Translation>);
            this.lisDefinitions();
          }
        })
        .catch();
    }
  };

  handleChangName = (value: string) => {
    const { list } = this.state;
    const newList: DefinitionBase[] = list.filter((item) => {
      return item.name && item.name.search(value) != -1;
    });
    this.setState({
      searchValue: value,
      searchList: newList,
      page: 1,
    });
  };

  getDataSource = () => {
    const { list, searchValue, searchList = [] } = this.state;
    if (!searchValue && searchList.length === 0) {
      return list;
    } else {
      return searchList;
    }
  };

  render() {
    const { definitionType, isLoading, searchValue, usageOf, pageSize } = this.state;
    const data = this.getDataSource();
    // A narrower search keeps the page in range.
    const page = Math.min(this.state.page, Math.max(1, Math.ceil(data.length / pageSize)));
    const columns = [
      {
        key: 'name',
        title: <Translation>Name</Translation>,
        dataIndex: 'name',
        cell: (v: string, i: number, record: DefinitionBase) => {
          return (
            <span className="definition-name">
              <Link to={`/definitions/${definitionType}/${v}/doc${definitionPlaceQuery(this.place(record))}`}>{v}</Link>
              {record.abstract && (
                <Tag size="small" className="definition-abstract">
                  <Translation>Abstract</Translation>
                </Tag>
              )}
              {record.extends && (
                <span className="definition-extends">
                  <Translation>extends</Translation> {record.extends}
                </span>
              )}
              {record.policyScope && <PolicyScopeTag scope={record.policyScope} />}
              {record.policy?.global && (
                // A global policy applies to every Application in its namespace, in priority order.
                <Balloon.Tooltip
                  align="t"
                  trigger={
                    <Tag size="small" className="definition-tag-help">
                      <Translation>Global</Translation>
                    </Tag>
                  }
                >
                  {i18n.t('Applied to every Application in its namespace')} ({i18n.t('priority')}{' '}
                  {record.policy.priority || 0})
                </Balloon.Tooltip>
              )}
            </span>
          );
        },
      },
      {
        key: 'where',
        title: <Translation>Where</Translation>,
        dataIndex: 'scope',
        cell: (v: string, i: number, record: DefinitionBase) => (
          <span>
            <StatusBadge
              tone={v === 'project' ? 'progressing' : 'neutral'}
              label={v === 'project' ? 'Project' : 'Global'}
            />
            {record.overridden && (
              <span className="definition-overridden">
                <Translation>{"overridden by the project's"}</Translation>
              </span>
            )}
          </span>
        ),
      },
      {
        key: 'status',
        title: <Translation>Status</Translation>,
        dataIndex: 'status',
        cell: (v: string) => {
          const enumStatusList = [
            { name: 'enable', color: 'enableStatus', status: 'Enabled' },
            { name: 'disable', color: 'disableStatus', status: 'Disabled' },
          ];
          const findStatus = _.find(enumStatusList, (item) => {
            return item.name === v;
          });
          return findStatus ? (
            <StatusBadge tone={findStatus.name === 'enable' ? 'healthy' : 'neutral'} label={findStatus.status} />
          ) : null;
        },
      },
      {
        key: 'restrictions',
        title: <Translation>Restrictions</Translation>,
        dataIndex: 'restrictions',
        cell: (v?: DefinitionRestrictions) => <RestrictionTags restrictions={v} />,
      },
      // {
      //   key: 'createTime',
      //   title: <Translation>Create Time</Translation>,
      //   dataIndex: 'createdTime',
      //   cell: (v: string) => {
      //     return <span>{momentDate(v)}</span>;
      //   },
      // },
      {
        key: 'operation',
        title: <Translation>Actions</Translation>,
        dataIndex: 'operation',
        lock: 'right',
        width: 160,
        cell: (v: string, i: number, record: DefinitionBase) => {
          const place = this.place(record);
          const project = place.where === 'project' ? this.project() : '';
          const may = (action: string) =>
            checkPermission({ resource: definitionResource(place, record.name), action }, project, this.props.userInfo);
          const enabled = record.status === 'enable';
          return (
            <RowLinks
              links={[
                {
                  label: enabled ? 'Disable' : 'Enable',
                  danger: enabled,
                  hidden: !may('update'),
                  onClick: () => this.onChangeStatus(record),
                },
                {
                  label: 'Usage',
                  hidden:
                    !(definitionType === 'component' || definitionType === 'trait') ||
                    (record.restrictions?.quota || []).length === 0 ||
                    !may('detail'),
                  onClick: () => this.setState({ usageOf: record }),
                },
              ]}
            />
          );
        },
      },
    ];

    const { Column } = Table;
    return (
      <div className="definitions-list-content console-panel">
        <div className="definitions-toolbar">
          <ListToolbar
            search={{
              value: searchValue,
              placeholder: 'Search by definition name',
              live: true,
              onSearch: (value: string) => this.handleChangName(value),
            }}
            onRefresh={() => this.lisDefinitions()}
            refreshing={isLoading}
          />
        </div>
        <Table
          id="definitionTable"
          locale={locale().Table}
          dataSource={pageOf(data, page, pageSize)}
          loading={isLoading}
          primaryKey="name"
        >
          {columns.map((col, key) => (
            <Column {...col} key={key} align={'left'} />
          ))}
        </Table>
        <TablePager
          total={data.length}
          page={page}
          pageSize={pageSize}
          onChange={(p, size) => this.setState({ page: p, pageSize: size })}
        />
        {usageOf && (definitionType === 'component' || definitionType === 'trait') && (
          <UsageDialog
            definition={usageOf}
            definitionType={definitionType}
            place={this.place(usageOf)}
            onClose={() => this.setState({ usageOf: undefined })}
          />
        )}
      </div>
    );
  }
}

export default Definitions;
