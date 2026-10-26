import { Table, Button, Message, Dialog } from '@alifd/next';
import { connect } from 'dva';
import { Link, routerRedux } from 'dva/router';
import querystring from 'query-string';
import React from 'react';

import { deployApplication } from '../../api/application';
import { notifyDeployed } from '../../utils/deploy';
import { listApplicationPods, listCloudResources } from '../../api/observation';
import { If } from '../../components/If';
import { Panel, RowLinks, TablePager, pageOf } from '../../components/Console';
import { StatusBadge } from '../../components/StatusBadge';
import type { Tone } from '../../components/StatusBadge';
import StatusShow from '../../components/StatusShow';
import { Translation } from '../../components/Translation';
import type {
  ApplicationComponent,
  ApplicationDeployResponse,
  ApplicationDetail,
  ApplicationStatus,
  EnvBinding,
  PodBase,
  CloudResource,
  Configuration,
  Target,
  LoginUserInfo,
} from '@velaux/data';
import { momentDate } from '../../utils/common';
import type { APIError } from '../../utils/errors';
import { handleError } from '../../utils/errors';
import { locale } from '../../utils/locale';

import Header from './components/Header';
import PodDetail from './components/PodDetail';

const { Column } = Table;
type Props = {
  dispatch: ({}) => void;
  match: {
    params: {
      envName: string;
      appName: string;
    };
  };
  applicationDetail?: ApplicationDetail;
  applicationStatus?: ApplicationStatus;
  envbinding: EnvBinding[];
  components?: ApplicationComponent[];
  userInfo?: LoginUserInfo;
};

type State = {
  podList?: PodBase[];
  cloudResourceList?: CloudResource[];
  envName: string;
  loading: boolean;
  target?: Target;
  componentName?: string;
  openRowKeys: [];
  cloudInstance?: CloudInstance[];
  showStatus: boolean;
  deployLoading: boolean;
  page: number;
  pageSize: number;
};

// podTones colours a pod's phase as the badges do.
const podTones: Record<string, Tone> = {
  Running: 'healthy',
  Pending: 'progressing',
  Succeeded: 'neutral',
  Failed: 'failed',
  Unknown: 'neutral',
};

type CloudInstance = {
  instanceName: string;
  status: string;
  region: string;
  createTime: string;
  message?: string;
  url?: string;
  type?: string;
};

@connect((store: any) => {
  return { ...store.application, ...store.user };
})
class ApplicationInstanceList extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    const { params } = props.match;
    this.state = {
      envName: params.envName,
      loading: false,
      openRowKeys: [],
      showStatus: false,
      deployLoading: false,
      page: 1,
      pageSize: 10,
    };
  }

  componentDidMount() {
    this.loadApplicationStatus();

    const query = querystring.parse(location.search);
    if (query && query.pod) {
      this.onRowOpen([query.pod as string]);
    }
  }

  componentWillReceiveProps(nextProps: any) {
    const { params } = nextProps.match;
    if (params.envName !== this.state.envName) {
      this.setState({ envName: params.envName }, () => {
        this.loadApplicationStatus();
      });
    }
  }

  loadApplicationStatus = async () => {
    const {
      params: { appName, envName },
    } = this.props.match;
    if (envName) {
      this.props.dispatch({
        type: 'application/getApplicationStatus',
        payload: { appName: appName, envName: envName },
        callback: (re: any) => {
          this.loadAppInstances();
        },
      });
    }
  };

  convertCloudInstance = (configurations: Configuration[]) => {
    const instances: CloudInstance[] = [];
    if (Array.isArray(configurations) && configurations.length > 0) {
      configurations.map((configuration) => {
        let url = '';
        let identifierKey = '';
        if (configuration.metadata.annotations) {
          url = configuration.metadata.annotations['cloud-resource/console-url'];
          identifierKey = configuration.metadata.annotations['cloud-resource/identifier'];
        }
        const outputs = configuration.status?.apply?.outputs;
        let instanceName = '';
        const region = configuration.status?.apply?.region || configuration.spec.region || '';
        if (outputs) {
          if (outputs[identifierKey]) {
            instanceName = outputs[identifierKey].value;
          }
          if (url) {
            const params = url.match(/\{(.+?)\}/g);
            if (Array.isArray(params) && params.length > 0) {
              params.map((param) => {
                const paramKey = param.replace('{', '').replace('}', '');
                if (paramKey.toLowerCase() == 'region' && region) {
                  url = url.replace(param, region);
                }
                if (outputs[paramKey]) {
                  url = url.replace(param, outputs[paramKey].value);
                }
              });
            }
          }
        }
        instances.push({
          instanceName: instanceName,
          status: configuration.status?.apply?.state || '-',
          url: url,
          createTime: configuration.metadata.creationTimestamp || '',
          region: region,
          message: configuration.status?.apply?.message,
          type: configuration.metadata.labels && configuration.metadata.labels['workload.oam.dev/type'],
        });
      });
    }
    this.setState({ cloudInstance: instances });
  };

  loadAppInstances = async () => {
    this.setState({ podList: [] });
    const { applicationDetail, envbinding, components } = this.props;
    const {
      params: { appName, envName },
    } = this.props.match;
    const cloudComponents = components?.filter((c) => c.workloadType?.type == 'configurations.terraform.core.oam.dev');
    const showCloudInstance = cloudComponents?.length && cloudComponents?.length > 0;
    const queryPod =
      cloudComponents?.length == undefined || (components?.length && components.length > cloudComponents?.length);
    const { target, componentName } = this.state;
    const envs = envbinding.filter((item) => item.name == envName);
    if (applicationDetail && applicationDetail.name && envs.length > 0) {
      const param = {
        project: applicationDetail?.project?.name || '',
        appName: envs[0].appDeployName || appName,
        appNs: envs[0].appDeployNamespace,
        componentName: componentName,
        cluster: '',
        clusterNs: '',
      };
      if (target) {
        param.cluster = target.cluster?.clusterName || '';
        param.clusterNs = target.cluster?.namespace || '';
      }
      this.setState({ loading: true });
      if (queryPod) {
        listApplicationPods(param)
          .then((re) => {
            if (re && re.podList) {
              re.podList.map((item: any) => {
                item.primaryKey = item.metadata.name;
              });
              this.setState({ podList: re.podList });
            } else {
              this.setState({ podList: [] });
            }
          })
          .finally(() => {
            this.setState({ loading: false });
          });
      }
      if (showCloudInstance) {
        this.setState({ loading: true });
        listCloudResources(param)
          .then((cloudResources: any) => {
            if (cloudResources) {
              this.convertCloudInstance(cloudResources['cloud-resources']);
            }
          })
          .finally(() => {
            this.setState({ loading: false });
          });
      } else {
        this.setState({ loading: false });
      }
    } else {
      setTimeout(() => {
        this.loadAppInstances();
      }, 1000);
    }
  };
  getColumns = () => {
    const { applicationDetail, components } = this.props;
    const targets = this.getTargets();
    const targetMap = new Map<string, Target>();
    targets?.map((item) => {
      targetMap.set(item.cluster?.clusterName + '-' + item.cluster?.namespace, item);
    });
    const getTarget = (key: string) => {
      return targetMap.get(key);
    };
    const componentNameAlias: any = {};
    components?.map((c) => {
      componentNameAlias[c.name] = c.alias || c.name;
    });
    return [
      {
        key: 'podName',
        title: <Translation>Pod Name</Translation>,
        dataIndex: 'metadata.name',
        cell: (v: string, index: number, record: PodBase) => {
          return (
            <div>
              <div>{record.metadata.name}</div>
              <div>{record.status.podIP}</div>
            </div>
          );
        },
      },
      {
        key: 'clusterName',
        title: <Translation>Target</Translation>,
        dataIndex: 'cluster',
        cell: (v: string, index: number, record: PodBase) => {
          const target = getTarget(record.cluster + '-' + record.metadata.namespace);
          if (target) {
            return <span>{target?.alias || target?.name}</span>;
          }
          return (
            <span>
              {record.cluster}/{record.metadata.namespace}
            </span>
          );
        },
      },
      {
        key: 'component',
        title: <Translation>Component</Translation>,
        dataIndex: 'component',
        cell: (v: string) => {
          return componentNameAlias[v];
        },
      },
      {
        key: 'status',
        title: <Translation>Status</Translation>,
        dataIndex: 'status.phase',
        cell: (v: string) => {
          return <StatusBadge tone={podTones[v] || 'neutral'} label={v || '-'} />;
        },
      },
      {
        key: 'createTime',
        title: <Translation>CreateTime</Translation>,
        dataIndex: 'metadata.creationTime',
        cell: (v: string) => {
          return <span>{momentDate(v)}</span>;
        },
      },
      {
        key: 'deployVersion',
        title: <Translation>Revision</Translation>,
        dataIndex: 'metadata.version.deployVersion',
        cell: (v: string) => {
          return (
            <span>
              <Link to={`/applications/${applicationDetail?.name}/revisions`}>{v}</Link>
            </span>
          );
        },
      },
      {
        key: 'workload',
        title: <Translation>Workload Type</Translation>,
        dataIndex: 'workload.kind',
        cell: (v: string) => {
          return <span>{v}</span>;
        },
      },
      {
        key: 'nodeName',
        title: <Translation>Node</Translation>,
        dataIndex: 'status.nodeName',
        cell: (v: string, index: number, record: PodBase) => {
          return (
            <div>
              <div>{record.status.nodeName}</div>
              <div>{record.status.hostIP}</div>
            </div>
          );
        },
      },
      {
        key: 'actions',
        title: <Translation>Actions</Translation>,
        dataIndex: 'primaryKey',
        width: 160,
        lock: 'right',
        cell: (v: string) => {
          const open = (this.state.openRowKeys as string[]).includes(v);
          return (
            <RowLinks
              links={[
                {
                  label: open ? 'Hide' : 'Details',
                  onClick: () => {
                    const keys = this.state.openRowKeys as string[];
                    this.onRowOpen(open ? keys.filter((k) => k !== v) : [...keys, v]);
                  },
                },
                {
                  label: 'Logs',
                  onClick: () =>
                    this.props.dispatch(
                      routerRedux.push(
                        `/applications/${applicationDetail?.name}/envbinding/${this.props.match.params.envName}/logs`
                      )
                    ),
                },
              ]}
            />
          );
        },
      },
    ];
  };

  updateQuery = (params: { target?: string; component?: string }) => {
    const targets = this.getTargets()?.filter((item) => item.name == params.target);
    let target: Target | undefined = undefined;
    if (targets && targets.length > 0) {
      target = targets[0];
    }
    this.setState({ target: target, componentName: params.component }, () => {
      this.loadAppInstances();
    });
  };

  getTargets = () => {
    const { envbinding, match } = this.props;
    const env = envbinding.filter((item) => item.name == match.params.envName);
    if (env.length > 0) {
      return env[0].targets;
    }
    return [];
  };

  onRowOpen = (openRowKeys: any) => {
    this.setState({ openRowKeys });
  };

  onDeploy = (force?: boolean) => {
    const { envbinding, dispatch } = this.props;
    const {
      params: { appName, envName },
    } = this.props.match;
    const envs = envbinding.filter((item) => item.name == envName);
    if (envs) {
      this.setState({ deployLoading: true });
      deployApplication(
        {
          appName: appName,
          workflowName: 'workflow-' + envs[0].name,
          triggerType: 'web',
          force: force || false,
        },
        true
      )
        .then((re: ApplicationDeployResponse) => {
          if (re) {
            notifyDeployed(re);
            this.setState({ deployLoading: false });
            this.loadApplicationStatus();
            if (re.record && re.record.name && dispatch) {
              dispatch(
                routerRedux.push(`/applications/${appName}/envbinding/${re.envName}/workflow/records/${re.record.name}`)
              );
            }
          }
        })
        .catch((err: APIError) => {
          if (err.BusinessCode === 10004) {
            Dialog.confirm({
              content: 'Workflow is executing. Do you want to force a restart?',
              onOk: () => {
                this.onDeploy(true);
              },
              onCancel: () => {
                this.setState({ deployLoading: false });
              },
              locale: locale().Dialog,
            });
          } else {
            handleError(err);
          }
        });
    } else {
      Message.warning('Please wait');
    }
  };
  getEnvbindingByName = () => {
    const { envbinding } = this.props;
    const {
      params: { envName },
    } = this.props.match;
    return envbinding.find((env) => env.name === envName);
  };

  loadStatusDetail = async () => {
    const {
      params: { appName, envName },
    } = this.props.match;
    if (envName) {
      this.setState({ loading: true });
      this.props.dispatch({
        type: 'application/getApplicationStatus',
        payload: { appName: appName, envName: envName },
        callback: () => {
          this.setState({ loading: false });
        },
      });
    }
  };

  onStatusClose = () => {
    this.setState({ showStatus: false });
  };

  render() {
    const { applicationStatus, applicationDetail, components, userInfo } = this.props;
    const { podList, loading, showStatus, cloudInstance, deployLoading, pageSize } = this.state;
    // A refresh that shrinks the list keeps the page in range.
    const page = Math.min(this.state.page, Math.max(1, Math.ceil((podList || []).length / pageSize)));
    const columns = this.getColumns();
    const envbinding = this.getEnvbindingByName();
    const expandedRowRender = (record: PodBase) => {
      return (
        <div style={{ margin: '16px 0' }}>
          <PodDetail
            env={envbinding}
            userInfo={userInfo}
            clusterName={record.cluster}
            application={applicationDetail}
            pod={record}
          />
        </div>
      );
    };
    const {
      params: { envName, appName },
    } = this.props.match;

    const cloudComponents = components?.filter((c) => c.workloadType?.type == 'configurations.terraform.core.oam.dev');
    const showCloudInstance = cloudComponents?.length && cloudComponents?.length > 0;

    const onlyShowCloudInstance = showCloudInstance && cloudComponents?.length == components?.length;
    return (
      <div>
        <Header
          envbinding={envbinding}
          userInfo={userInfo}
          targets={this.getTargets()}
          components={components}
          envName={envName}
          appName={appName}
          applicationDetail={applicationDetail}
          applicationStatus={applicationStatus}
          updateQuery={(params: { target?: string; component?: string }) => {
            this.updateQuery(params);
          }}
          refresh={() => {
            this.loadApplicationStatus();
          }}
          dispatch={this.props.dispatch}
        />
        <If condition={applicationStatus}>
          <If condition={!onlyShowCloudInstance}>
            <div className="console-panel pod-panel">
              <Table
                locale={locale().Table}
                className="podlist-table-wrapper"
                size="medium"
                primaryKey={'primaryKey'}
                loading={loading}
                dataSource={pageOf(podList || [], page, pageSize)}
                expandedIndexSimulate
                expandedRowRender={expandedRowRender}
                openRowKeys={this.state.openRowKeys}
                onRowOpen={(openRowKeys: any) => {
                  this.onRowOpen(openRowKeys);
                }}
              >
                {columns.map((col) => (
                  <Column {...col} key={col.key} align={'left'} />
                ))}
              </Table>
              <TablePager
                total={(podList || []).length}
                page={page}
                pageSize={pageSize}
                onChange={(p, size) => this.setState({ page: p, pageSize: size })}
              />
            </div>
          </If>
          <If condition={showCloudInstance}>
            <Panel title="Instances of the cloud service" flush>
              <div>
                <Table
                  size="medium"
                  locale={locale().Table}
                  className="customTable"
                  dataSource={cloudInstance}
                  primaryKey={'instanceName'}
                  loading={loading}
                >
                  <Column
                    align="left"
                    title={<Translation>Name</Translation>}
                    dataIndex="instanceName"
                    cell={(value: string, index: number, record: CloudInstance) => {
                      if (record.url) {
                        return (
                          <a target="_blank" href={record.url} rel="noopener noreferrer">
                            {value}
                          </a>
                        );
                      }
                      return value;
                    }}
                  />
                  <Column
                    align="left"
                    title={<Translation>Status</Translation>}
                    dataIndex="status"
                    cell={(v: string) => <StatusBadge tone={v ? 'neutral' : 'undeployed'} label={v || '-'} />}
                  />
                  <Column align="left" title={<Translation>Resource Type</Translation>} dataIndex="type" />
                  <Column
                    align="left"
                    title={<Translation>Create Time</Translation>}
                    dataIndex="createTime"
                    cell={(v: string) => {
                      return <span>{momentDate(v)}</span>;
                    }}
                  />
                  <Column align="left" title={<Translation>Region</Translation>} dataIndex="region" />
                  <Column
                    align="left"
                    title={<Translation>Actions</Translation>}
                    dataIndex="url"
                    width={120}
                    lock="right"
                    cell={(value: string, index: number, record: CloudInstance) => {
                      if (record.instanceName) {
                        return (
                          <RowLinks
                            links={[
                              {
                                label: 'Console',
                                disabled: !value,
                                onClick: () => window.open(value, '_blank', 'noopener,noreferrer'),
                              },
                            ]}
                          />
                        );
                      }
                      return;
                    }}
                  />
                </Table>
              </div>
            </Panel>
          </If>
        </If>
        <If condition={!applicationStatus}>
          <div className="deployNotice">
            <div className="noticeBox">
              <h2>
                <Translation>Not Deploy</Translation>
              </h2>
              <div className="desc">
                <Translation>The current environment has not been deployed.</Translation>
              </div>
              <div className="noticeAction">
                <Button
                  loading={deployLoading}
                  disabled={applicationDetail?.readOnly}
                  onClick={() => this.onDeploy()}
                  type="primary"
                >
                  <Translation>Deploy</Translation>
                </Button>
              </div>
            </div>
          </div>
        </If>
        <If condition={showStatus}>
          <StatusShow
            loading={loading}
            title={<Translation>Application Status</Translation>}
            applicationStatus={applicationStatus}
            loadStatusDetail={this.loadStatusDetail}
            onClose={this.onStatusClose}
          />
        </If>
      </div>
    );
  }
}

export default ApplicationInstanceList;
