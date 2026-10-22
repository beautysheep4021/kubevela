import { ListToolbar, PageGuide, RowLinks, TablePager, pageOf } from '../../components/Console';

import { projectChanged } from '../../utils/currentProject';
import { Loading, Button, Table, Dialog, Message, Balloon } from '@alifd/next';
import { connect } from 'dva';
import { Link, routerRedux } from 'dva/router';
import React, { Component } from 'react';
import { RelativeTime } from '../../components/RelativeTime';
import { AiOutlineHistory, AiOutlinePartition, AiOutlinePlusSquare } from 'react-icons/ai';
import type { Dispatch } from 'redux';

import { deletePipeline, listPipelines } from '../../api/pipeline';

import './index.less';
import { If } from '../../components/If';
import Permission from '../../components/Permission';
import RunPipeline from '../../components/RunPipeline';
import { Translation } from '../../components/Translation';
import i18n from '../../i18n';
import type {
  AddonBaseStatus,
  NameAlias,
  PipelineBase,
  PipelineListItem,
  PipelineRun,
  RunStateInfo,
  LoginUserInfo,
} from '@velaux/data';
import { momentDate } from '../../utils/common';
import { locale } from '../../utils/locale';
import { checkPermission } from '../../utils/permission';
import type { Tone } from '../../components/StatusBadge';
import { StatusBadge } from '../../components/StatusBadge';

import CreatePipeline from './components/CreatePipeline';
import ClonePipeline from './components/PipelineClone';
import ViewRuns from './components/ViewRuns';

type Props = {
  userInfo?: LoginUserInfo;
  dispatch: Dispatch<any>;
  enabledAddons?: AddonBaseStatus[];
  currentProject?: { current: string; resolved: boolean };
};

export type ShowMode = 'table' | 'card' | string | null;

type State = {
  isLoading: boolean;
  editItem?: PipelineListItem;
  pipelines?: PipelineListItem[];
  query: string;
  page: number;
  pageSize: number;
  showRunPipeline?: boolean;
  pipeline?: PipelineListItem;
  showRuns?: boolean;
  showNewPipeline?: boolean;
  showClonePipeline?: boolean;
};

@connect((store: any) => {
  return { ...store.user, ...store.addons, currentProject: store.currentProject };
})
class PipelineListPage extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      isLoading: false,
      query: '',
      page: 1,
      pageSize: 10,
    };
  }

  componentDidMount() {
    this.getPipelines({});
  }

  componentDidUpdate(prev: Props) {
    if (projectChanged(prev.currentProject, this.props.currentProject)) {
      this.getPipelines({});
    }
  }

  // getPipelines lists the picked project's pipelines, or every project's for all of them.
  getPipelines = async (params: { query?: string }) => {
    if (!this.props.currentProject?.resolved) {
      return;
    }
    const query = params.query ?? this.state.query;
    this.setState({ isLoading: true, query, page: 1 });
    listPipelines({ query, projectName: this.props.currentProject.current })
      .then((res) => {
        this.setState({
          pipelines: res && Array.isArray(res.pipelines) ? res.pipelines : [],
        });
      })
      .finally(() => {
        this.setState({ isLoading: false });
      });
  };

  onShowPipelineRuns = (pipeline: PipelineListItem) => {
    this.setState({ showRuns: true, pipeline: pipeline });
  };

  onRunPipeline = (pipeline: PipelineListItem) => {
    this.setState({ pipeline: pipeline, showRunPipeline: true });
  };

  onClonePipeline = (pipeline: PipelineListItem) => {
    this.setState({ pipeline, showClonePipeline: true });
  };

  onDeletePipeline = (pipeline: PipelineListItem) => {
    Dialog.confirm({
      type: 'confirm',
      content: <Translation>Unrecoverable after deletion, are you sure to delete it?</Translation>,
      onOk: () => {
        deletePipeline({
          projectName: pipeline.project.name,
          pipelineName: pipeline.name,
        }).then((res) => {
          if (res) {
            Message.success(i18n.t('The Pipeline removed successfully'));
            this.getPipelines({});
          }
        });
      },
      locale: locale().Dialog,
    });
  };

  onEditPipeline = (pipeline: PipelineListItem) => {
    this.setState({ showNewPipeline: true, pipeline: pipeline });
  };

  // rowLinks are a pipeline's operations, each shown to those who may use it.
  rowLinks = (pipeline: PipelineListItem) => {
    const project = pipeline.project.name;
    const may = (resource: string, action: string) =>
      checkPermission({ resource: `project:${project}/${resource}`, action }, project, this.props.userInfo);
    return (
      <RowLinks
        links={[
          {
            label: 'Run',
            hidden: !may(`pipeline:${pipeline.name}`, 'run'),
            onClick: () => this.onRunPipeline(pipeline),
          },
          {
            label: 'View Runs',
            hidden: !may(`pipeline:${pipeline.name}/pipelineRun:*`, 'list'),
            onClick: () => this.onShowPipelineRuns(pipeline),
          },
          {
            label: 'Edit',
            hidden: !may(`pipeline:${pipeline.name}`, 'update'),
            onClick: () => this.onEditPipeline(pipeline),
          },
          {
            label: 'Clone',
            hidden: !may('pipeline:*', 'create'),
            onClick: () => this.onClonePipeline(pipeline),
          },
          {
            label: 'Remove',
            danger: true,
            hidden: !may(`pipeline:${pipeline.name}`, 'delete'),
            onClick: () => this.onDeletePipeline(pipeline),
          },
        ]}
      />
    );
  };

  renderPipelineTable = () => {
    const pipelines = this.state.pipelines || [];
    const { pageSize } = this.state;
    const page = Math.min(this.state.page, Math.max(1, Math.ceil(pipelines.length / pageSize)));
    return (
      <div className="table-pipeline-list">
        <Table
          className="customTable"
          size="medium"
          locale={locale().Table}
          dataSource={pageOf(pipelines, page, pageSize)}
          primaryKey="name"
        >
          <Table.Column
            key={'name'}
            title={i18n.t('Name(Alias)').toString()}
            dataIndex="name"
            cell={(name: string, i: number, pipeline: PipelineListItem) => {
              let text = name;
              if (pipeline.alias) {
                text = `${name}(${pipeline.alias})`;
              }
              return (
                <div className="pipeline-name">
                  <Link to={`/projects/${pipeline.project.name}/pipelines/${pipeline.name}/studio`}>{text}</Link>
                  <span>{pipeline.description}</span>
                </div>
              );
            }}
          />
          <Table.Column
            key={'project'}
            title={i18n.t('Project').toString()}
            dataIndex="project"
            cell={(project: NameAlias) => {
              let text = project.name;
              if (project.alias) {
                text = `${project.name}(${project.alias})`;
              }
              return <Link to={`/projects/${project.name}`}>{text}</Link>;
            }}
          />
          <Table.Column
            key={'createTime'}
            title={i18n.t('CreateTime').toString()}
            dataIndex="createTime"
            width={180}
            cell={(v: string) => {
              return momentDate(v);
            }}
          />
          <Table.Column
            key={'runs'}
            title={i18n.t('Recent Runs(Last 7-Days)').toString()}
            dataIndex="info.runStat"
            width={'280px'}
            cell={(runState: { activeNum: number; total: RunStateInfo; week?: RunStateInfo[] }) => {
              if (!runState) {
                return '-';
              }
              return (
                <div className="run-state">
                  <div className="week">
                    {runState.week?.map((day) => {
                      const failure = day.total == 0 ? 0 : Math.floor((day.fail / day.total) * 100);
                      const success = 100 - failure;
                      return (
                        <Balloon
                          trigger={
                            <div className="rectangle">
                              <If condition={day.total > 0}>
                                <span className="failure" style={{ height: `${failure}%` }} />
                                <span className="success" style={{ height: `${success}%` }} />
                              </If>
                              <If condition={day.total == 0}>
                                <span style={{ height: `10px` }} />
                              </If>
                            </div>
                          }
                        >
                          <If condition={day.total == 0}>No Run</If>
                          <If condition={day.total > 0}>
                            <p>Total: {day.total}</p>
                            <p>Success: {day.success}</p>
                          </If>
                        </Balloon>
                      );
                    })}
                  </div>
                  <If condition={runState.total.total > 0}>
                    <div className="active">
                      <span>{runState.total.total}</span>
                    </div>
                  </If>
                </div>
              );
            }}
          />
          <Table.Column
            key={'lastRun'}
            title={i18n.t('Last Run').toString()}
            dataIndex="info.lastRun"
            cell={(run?: PipelineRun) => {
              if (run) {
                return (
                  <div className="last-run">
                    <div className="metadata">
                      <Link
                        to={`/projects/${run.project.name}/pipelines/${run.pipelineName}/runs/${run.pipelineRunName}`}
                      >
                        {run.pipelineRunName}
                      </Link>
                      <RelativeTime time={run.status?.startTime} />
                    </div>
                    <StatusBadge tone={runTone(run.status?.status)} label={runLabel(run.status?.status)} />
                  </div>
                );
              }
              return <span style={{ color: 'var(--grey-400)' }}>This Pipeline never ran.</span>;
            }}
          />
          <Table.Column
            key={'actions'}
            title={i18n.t('Actions').toString()}
            dataIndex="name"
            width={200}
            lock="right"
            cell={(name: string, i: number, pipeline: PipelineListItem) => this.rowLinks(pipeline)}
          />
        </Table>
        <TablePager
          total={pipelines.length}
          page={page}
          pageSize={pageSize}
          onChange={(p, size) => this.setState({ page: p, pageSize: size })}
        />
      </div>
    );
  };

  render() {
    const { query, isLoading, showRunPipeline, pipeline, showRuns, showNewPipeline, showClonePipeline } = this.state;
    const { enabledAddons } = this.props;
    const addonEnabled = enabledAddons?.filter((addon) => addon.name == 'vela-workflow').length;
    return (
      <div>
        <PageGuide
          id="pipelines"
          title="Pipelines"
          description="Orchestrate your multiple cloud native app release processes or model your cloud native pipeline."
          steps={[
            {
              icon: <AiOutlinePlusSquare />,
              title: 'Create a pipeline',
              description: <Translation>Name it in a project, or clone one that exists</Translation>,
            },
            {
              icon: <AiOutlinePartition />,
              title: 'Configure its steps',
              description: <Translation>Lay out its steps, their inputs and outputs in the studio</Translation>,
            },
            {
              icon: <AiOutlineHistory />,
              title: 'Run and watch',
              description: <Translation>Run it with a context, then follow each run step by step</Translation>,
            },
          ]}
        />

        <If condition={addonEnabled}>
          <div className="console-panel">
            <div className="console-panel-body pipeline-toolbar">
              <ListToolbar
                search={{
                  value: query,
                  placeholder: 'Search by Name and Description etc',
                  onSearch: (value: string) => this.getPipelines({ query: value }),
                }}
                onRefresh={() => this.getPipelines({})}
                refreshing={isLoading}
                actions={
                  <Permission request={{ resource: 'project:?/pipeline:*', action: 'create' }} project={'?'}>
                    <Button
                      type="primary"
                      onClick={() => {
                        this.setState({ showNewPipeline: true });
                      }}
                    >
                      <Translation>New Pipeline</Translation>
                    </Button>
                  </Permission>
                }
              />
            </div>
            <Loading style={{ width: '100%' }} visible={isLoading}>
              {this.renderPipelineTable()}
            </Loading>
          </div>
        </If>
        <If condition={!addonEnabled}>
          <Permission request={{ resource: 'addon', action: 'enable' }}>
            <div className="addon-notice">
              Please enable the <Link to="/addons/vela-workflow">vela-workflow</Link> addon to power pipeline.
            </div>
          </Permission>
        </If>

        <If condition={showRunPipeline}>
          {pipeline && (
            <RunPipeline
              onClose={() => {
                this.setState({ showRunPipeline: false, pipeline: undefined });
              }}
              onSuccess={(runName) => {
                this.props.dispatch(
                  routerRedux.push(`/projects/${pipeline.project.name}/pipelines/${pipeline.name}/runs/${runName}`)
                );
              }}
              pipeline={pipeline}
            />
          )}
        </If>

        <If condition={showRuns}>
          {pipeline && (
            <ViewRuns
              pipeline={pipeline}
              onClose={() => {
                this.setState({ showRuns: false, pipeline: undefined });
              }}
            />
          )}
        </If>
        <If condition={showNewPipeline}>
          <CreatePipeline
            project={this.props.currentProject?.current || undefined}
            onClose={() => {
              this.setState({ showNewPipeline: false, pipeline: undefined });
            }}
            onSuccess={(p: PipelineBase) => {
              this.props.dispatch(routerRedux.push(`/projects/${p.project.name}/pipelines/${p.name}/studio`));
            }}
            pipeline={pipeline}
          />
        </If>

        <If condition={showClonePipeline}>
          <ClonePipeline
            onClose={() => {
              this.setState({ showClonePipeline: false, pipeline: undefined });
            }}
            onSuccess={() => {
              this.getPipelines({});
              this.setState({ showClonePipeline: false, pipeline: undefined });
            }}
            pipeline={pipeline}
          />
        </If>
      </div>
    );
  }
}

export default PipelineListPage;

// runTone colours a pipeline run by its phase.
function runTone(phase?: string): Tone {
  switch (phase) {
    case 'succeeded':
      return 'healthy';
    case 'failed':
    case 'terminated':
      return 'failed';
    case 'suspending':
      return 'suspended';
    case 'executing':
    case 'initializing':
      return 'progressing';
    default:
      return 'neutral';
  }
}

// runLabel writes a run's phase as a word: "succeeded" to "Succeeded".
function runLabel(phase?: string): string {
  return phase ? phase.charAt(0).toUpperCase() + phase.slice(1) : 'Unknown';
}
