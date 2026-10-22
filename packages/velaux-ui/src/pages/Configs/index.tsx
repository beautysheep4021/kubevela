import { Table, Button, Dialog, Message, Tag, Balloon, Tab } from '@alifd/next';
import { connect } from 'dva';
import { routerRedux } from 'dva/router';
import React, { Component } from 'react';
import { AiOutlineAppstore, AiOutlineForm, AiOutlineSend } from 'react-icons/ai';

import { getConfigs, getProjectConfigs, deleteConfig, listTemplates } from '../../api/config';
import Empty from '../../components/Empty';
import { ListToolbar, PageGuide, RowLinks, TablePager, pageOf } from '../../components/Console';
import Permission from '../../components/Permission';
import { StatusBadge } from '../../components/StatusBadge';
import { Translation } from '../../components/Translation';
import i18n from '../../i18n';
import type { ConfigTemplate, Config, LoginUserInfo } from '@velaux/data';
import { momentDate } from '../../utils/common';
import { allProjects, projectChanged } from '../../utils/currentProject';
import { locale } from '../../utils/locale';
import { checkPermission } from '../../utils/permission';
import DistributeConfigDialog from '../ProjectSummary/components/Configs/config-distribute';
import CreateConfigDialog from './components/CreateConfigDialog';
import { templateIcon } from './icons';
import { templateOptions } from './templates';
import './index.less';

type Props = {
  userInfo?: LoginUserInfo;
  location?: { search?: string };
  dispatch: ({}) => {};
  // scope project lists the configs of the project picked in the top bar, and
  // the global ones shared with it; without it, the global configs.
  scope?: 'project';
  currentProject?: { current: string; resolved: boolean };
};

type State = {
  list: Config[];
  templates: ConfigTemplate[];
  // template filters the list to one template's configs; empty is all.
  template: string;
  query: string;
  // open is the config the dialog shows, or new for a config not yet made.
  open?: { name?: string; template?: string; templateNamespace?: string };
  // distribution is the legacy project config being distributed to targets.
  distribution?: Config;
  isLoading: boolean;
  page: number;
  pageSize: number;
};

// templateParam is the template a link to the list filters it to.
function templateParam(search?: string): string {
  return new URLSearchParams(search || '').get('template') || '';
}

@connect((store: any) => {
  return { ...store.user, currentProject: store.currentProject };
})
class Configs extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      list: [],
      templates: [],
      template: templateParam(props.location?.search),
      query: '',
      isLoading: false,
      page: 1,
      pageSize: 10,
    };
  }

  componentDidMount() {
    this.load();
  }

  componentDidUpdate(prevProps: Props) {
    if (this.props.scope === 'project' && projectChanged(prevProps.currentProject, this.props.currentProject)) {
      this.load();
    }
  }

  // project is the project whose configs the page lists: in project scope the
  // one picked in the top bar, else none, the global configs.
  project = () =>
    this.props.scope === 'project' && this.props.currentProject?.resolved ? this.props.currentProject.current : '';

  load = () => {
    if (this.props.scope === 'project' && !this.project()) {
      this.setState({ list: [], templates: [] });
      return;
    }
    this.listConfigs();
    listTemplates(this.project() || undefined).then((res) => {
      this.setState({ templates: (res && res.templates) || [] });
    });
  };

  listConfigs = () => {
    const project = this.project();
    this.setState({ isLoading: true });
    (project ? getProjectConfigs({ projectName: project }) : getConfigs())
      .then((res: any) => {
        this.setState({ list: (res && res.configs) || [] });
      })
      .finally(() => {
        this.setState({ isLoading: false });
      });
  };

  // setTemplate filters the list, keeping the filter in the URL so it can be linked to.
  setTemplate = (template: string) => {
    const path = this.props.scope === 'project' ? '/project-configs' : '/configs';
    this.setState({ template, page: 1 });
    this.props.dispatch(routerRedux.replace(template ? `${path}?template=${template}` : path));
  };

  onDelete = (record: Config) => {
    Dialog.confirm({
      content: 'Are you sure want to delete this config',
      onOk: () => {
        deleteConfig(record.name, this.project() || undefined).then((res) => {
          if (res) {
            Message.success(<Translation>Config deleted successfully</Translation>);
            this.listConfigs();
          }
        });
      },
      locale: locale().Dialog,
    });
  };

  onSuccess = () => {
    this.setState({ open: undefined });
    this.listConfigs();
  };

  templateOf = (name: string) => this.state.templates.find((t) => t.name === name);

  // rowLinks are a config's operations. In a project a shared config is the
  // platform's, changed on the global page, so it has none here.
  rowLinks = (record: Config) => {
    const { userInfo } = this.props;
    const inProject = this.props.scope === 'project';
    const project = this.project();
    const may = (action: string) =>
      inProject
        ? checkPermission({ resource: `project:${project}/config:${record.name}`, action }, project, userInfo)
        : checkPermission({ resource: `config:${record.name}`, action }, '', userInfo);
    if (inProject && record.shared) {
      return <span className="row-list-muted">-</span>;
    }
    const viewable = !record.sensitive;
    return (
      <RowLinks
        links={[
          {
            label: 'Details',
            disabled: !viewable,
            reason: 'Sensitive configs cannot be viewed',
            onClick: () =>
              this.setState({
                open: {
                  name: record.name,
                  template: record.template?.name,
                  templateNamespace: record.template?.namespace,
                },
              }),
          },
          {
            label: 'Distribute',
            hidden: !inProject || !record.legacy || !may('distribute'),
            onClick: () => this.setState({ distribution: record }),
          },
          { label: 'Delete', danger: true, hidden: !may('delete'), onClick: () => this.onDelete(record) },
        ]}
      />
    );
  };

  render() {
    const { list, templates, template, query, open, isLoading, distribution } = this.state;
    const inProject = this.props.scope === 'project';
    const project = this.project();
    const q = query.trim().toLowerCase();
    const matched = list.filter(
      (c) =>
        !q ||
        [c.name, c.alias, c.description, c.template?.name, c.templateAlias].some((v) =>
          (v || '').toLowerCase().includes(q)
        )
    );
    const shown = matched.filter((c) => !template || c.template?.name === template);
    const { pageSize } = this.state;
    // A filter that shrinks the list keeps the page in range.
    const page = Math.min(this.state.page, Math.max(1, Math.ceil(shown.length / pageSize)));
    // The tabs are the templates with configs, and the one the URL picks,
    // each counting its configs the search leaves.
    const tabs = templateOptions(templates, list).filter(
      (o) => o.name === template || list.some((c) => c.template?.name === o.name)
    );
    const countOf = (name: string) => matched.filter((c) => c.template?.name === name).length;
    const templateTab = (t: ConfigTemplate | { name: string; alias?: string }, count: number) => (
      <span className="config-template">
        <img src={templateIcon(t.name)} />
        {t.alias || t.name}
        <span className="configs-tab-count">{count}</span>
      </span>
    );
    const columns = [
      {
        key: 'name',
        title: <Translation>Name</Translation>,
        dataIndex: 'name',
        cell: (v: string, i: number, config: Config) =>
          config.sensitive || (inProject && config.shared) ? (
            <span>{v}</span>
          ) : (
            <a
              onClick={() =>
                this.setState({
                  open: { name: v, template: config.template?.name, templateNamespace: config.template?.namespace },
                })
              }
            >
              {v}
            </a>
          ),
      },
      {
        key: 'alias',
        title: <Translation>Alias</Translation>,
        dataIndex: 'alias',
        cell: (v: string) => v || '',
      },
      ...(inProject
        ? [
            {
              key: 'where',
              title: <Translation>Where</Translation>,
              dataIndex: 'shared',
              cell: (v: boolean) => (
                <StatusBadge tone={v ? 'neutral' : 'progressing'} label={v ? 'Global' : 'Project'} />
              ),
            },
          ]
        : []),
      {
        key: 'template',
        title: <Translation>Template</Translation>,
        dataIndex: 'template',
        cell: (v: Config['template'], i: number, config: Config) => {
          const t = this.templateOf(v?.name);
          return (
            <span className="config-template">
              <img src={templateIcon(v?.name || '')} />
              {config.templateAlias || t?.alias || v?.name}
              {config.legacy && (
                <Tag size="small">
                  <Translation>Legacy</Translation>
                </Tag>
              )}
            </span>
          );
        },
      },
      {
        key: 'phase',
        title: <Translation>Status</Translation>,
        dataIndex: 'phase',
        cell: (v: string, i: number, config: Config) => {
          if (config.legacy) {
            return <span>-</span>;
          }
          const tone = v === 'Available' ? 'healthy' : v === 'Error' ? 'failed' : 'progressing';
          const tag = (
            <span>
              <StatusBadge tone={tone} label={v || 'Pending'} />
            </span>
          );
          return config.message ? <Balloon.Tooltip trigger={tag}>{config.message}</Balloon.Tooltip> : tag;
        },
      },
      {
        key: 'description',
        title: <Translation>Description</Translation>,
        dataIndex: 'description',
      },
      {
        key: 'createTime',
        title: <Translation>Create Time</Translation>,
        dataIndex: 'createdTime',
        cell: (v: string) => <span>{momentDate(v)}</span>,
      },
      {
        key: 'operation',
        title: <Translation>Actions</Translation>,
        dataIndex: 'operation',
        width: 180,
        lock: 'right',
        cell: (v: string, i: number, record: Config) => this.rowLinks(record),
      },
    ];
    const { Column } = Table;
    return (
      <div className="list-content configs-list">
        <PageGuide
          id={inProject ? 'project-configs' : 'configs'}
          title="Configs"
          description={
            inProject
              ? "The project's configs, and the global ones shared with it"
              : 'Offering templated and extensible configuration management capabilities.'
          }
          steps={[
            {
              icon: <AiOutlineAppstore />,
              title: 'Pick a template',
              description: (
                <Translation>Choose what the config is for, such as an image registry or Helm repository</Translation>
              ),
            },
            {
              icon: <AiOutlineForm />,
              title: 'Fill its properties',
              description: (
                <Translation>Enter the values the template asks for; sensitive ones stay hidden</Translation>
              ),
            },
            {
              icon: <AiOutlineSend />,
              title: 'Distribute it',
              description: (
                <Translation>Share it with projects, or distribute it to clusters where it is used</Translation>
              ),
            },
          ]}
        />
        {inProject && project === allProjects ? (
          <Empty
            message={<Translation>Pick a project in the top bar. Each project has configs of its own.</Translation>}
          />
        ) : (
          <div className="console-panel">
            <div className="console-panel-body configs-head">
              <ListToolbar
                search={{
                  value: query,
                  placeholder: 'Search by name, alias, template or description',
                  live: true,
                  onSearch: (value: string) => this.setState({ query: value, page: 1 }),
                }}
                onRefresh={() => this.listConfigs()}
                refreshing={isLoading}
                actions={
                  <Permission
                    key="new"
                    request={{ resource: inProject ? `project:${project}/config:*` : `config:*`, action: 'create' }}
                    project={inProject ? project : ''}
                  >
                    <Button type="primary" onClick={() => this.setState({ open: { template: template || undefined } })}>
                      <Translation>New Config</Translation>
                    </Button>
                  </Permission>
                }
              />
              <Tab
                shape="pure"
                excessMode="slide"
                className="configs-tabs"
                activeKey={template || ''}
                onChange={(key: string | number) => this.setTemplate(String(key))}
              >
                <Tab.Item
                  key=""
                  title={
                    <span>
                      {i18n.t('All').toString()}
                      <span className="configs-tab-count">{matched.length}</span>
                    </span>
                  }
                />
                {tabs.map((t) => (
                  <Tab.Item key={t.name} title={templateTab(t, countOf(t.name))} />
                ))}
              </Tab>
            </div>
            <Table
              locale={locale().Table}
              dataSource={pageOf(shown, page, pageSize)}
              loading={isLoading}
              primaryKey="name"
            >
              {columns.map((col, key) => (
                <Column {...col} key={key} align={'left'} />
              ))}
            </Table>
            <TablePager
              total={shown.length}
              page={page}
              pageSize={pageSize}
              onChange={(p, size) => this.setState({ page: p, pageSize: size })}
            />
          </div>
        )}

        {open && (
          <CreateConfigDialog
            visible={true}
            project={project || undefined}
            configName={open.name}
            template={open.template ? { name: open.template, namespace: open.templateNamespace } : undefined}
            onSuccess={this.onSuccess}
            onClose={() => this.setState({ open: undefined })}
          />
        )}
        {distribution && (
          <DistributeConfigDialog
            config={distribution}
            targets={distribution.targets}
            projectName={project}
            onClose={() => this.setState({ distribution: undefined })}
            onSuccess={() => {
              this.setState({ distribution: undefined });
              this.listConfigs();
            }}
          />
        )}
      </div>
    );
  }
}

export default Configs;
