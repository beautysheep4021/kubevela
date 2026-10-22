import { Button, Table } from '@alifd/next';
import { connect } from 'dva';
import { Link, routerRedux } from 'dva/router';
import React, { useState } from 'react';
import { RelativeTime } from '../../components/RelativeTime';
import { AiOutlineDown, AiOutlineEnvironment, AiOutlineEye, AiOutlineRight } from 'react-icons/ai';

import type { ApplicationDetail, EnvBinding } from '@velaux/data';
import Empty from '../../components/Empty';
import { If } from '../../components/If';
import { ListToolbar, RowLinks } from '../../components/Console';
import { locale } from '../../utils/locale';
import Permission from '../../components/Permission';
import { Chip } from '../../components/ResourceCard';
import { RowAction } from '../../components/RowAction';
import '../../components/RowList';
import { StatusBadge } from '../../components/StatusBadge';
import { Translation } from '../../components/Translation';
import AddAndEditEnvBind from '../../layout/Application/components/AddAndEditEnvBind';
import type { EnvironmentStatus } from '../ApplicationList/components/AppStatus/health';
import {
  healthLabels,
  pausedEnvs,
  summariseStatuses,
  workflowLabel,
} from '../ApplicationList/components/AppStatus/health';
import './index.less';

type Props = {
  applicationDetail?: ApplicationDetail;
  envbinding?: EnvBinding[];
  applicationAllStatus?: EnvironmentStatus[];
  dispatch: (action: any) => void;
};

// ApplicationEnvironments lists the envs an application is bound to, each opening the
// env's live view.
const ApplicationEnvironments = (props: Props) => {
  const { applicationDetail, envbinding = [], applicationAllStatus = [], dispatch } = props;
  const [adding, setAdding] = useState(false);
  const appName = applicationDetail?.name || '';
  const projectName = applicationDetail?.project?.name || '';
  const reload = () => {
    dispatch({ type: 'application/getApplicationEnvbinding', payload: { appName } });
    dispatch({ type: 'application/getApplicationWorkflows', payload: { appName } });
    dispatch({ type: 'application/getApplicationPolicies', payload: { appName } });
    dispatch({ type: 'application/getApplicationAllStatus', payload: { appName } });
  };
  return (
    <div className="console-panel app-tab-panel">
      <ListToolbar
        filters={
          <span className="app-tab-hint">
            <Translation>Where the application deploys. Open one for its live status.</Translation>
          </span>
        }
        onRefresh={reload}
        actions={
          <If condition={!applicationDetail?.readOnly}>
            <Permission
              request={{ resource: `project:${projectName}/application:${appName}/envBinding:*`, action: 'create' }}
              project={projectName}
            >
              <Button type="primary" onClick={() => setAdding(true)}>
                <Translation>Add Environment</Translation>
              </Button>
            </Permission>
          </If>
        }
      />
      <EnvironmentTable
        appName={appName}
        envbinding={envbinding}
        applicationAllStatus={applicationAllStatus}
        dispatch={dispatch}
      />
      <If condition={adding}>
        <AddAndEditEnvBind
          envbinding={envbinding}
          onClose={() => setAdding(false)}
          onOK={() => {
            reload();
            setAdding(false);
          }}
        />
      </If>
    </div>
  );
};

export default connect((store: any) => ({ ...store.application }))(ApplicationEnvironments);

// EnvironmentTable lists the envs an application is bound to as a table of
// their health, each row expanding to its targets; its name and Open lead to
// the env's live view.
const EnvironmentTable = (props: {
  appName: string;
  envbinding: EnvBinding[];
  applicationAllStatus: EnvironmentStatus[];
  dispatch: (action: any) => void;
}) => {
  const { appName, envbinding, applicationAllStatus, dispatch } = props;
  const summary = summariseStatuses(applicationAllStatus);
  const paused = pausedEnvs(applicationAllStatus);
  const base = (env: string) => `/applications/${appName}/envbinding/${env}`;
  return (
    <Table
      dataSource={envbinding}
      primaryKey="name"
      locale={locale().Table}
      emptyContent={<Empty message={<Translation>This application is not bound to an environment yet</Translation>} />}
      expandedRowRender={(binding: EnvBinding) => <EnvironmentDetail binding={binding} />}
      expandedRowIndent={[1, 0]}
    >
      <Table.Column
        dataIndex="name"
        title={<Translation>Name</Translation>}
        cell={(v: string, i: number, binding: EnvBinding) => (
          <span className="environment-table-name">
            <Link to={`${base(v)}/status`}>{binding.alias || v}</Link>
            {binding.alias && binding.alias !== v && <span className="row-list-type">{v}</span>}
          </span>
        )}
      />
      <Table.Column
        dataIndex="name"
        title={<Translation>Health</Translation>}
        cell={(v: string) => {
          const health = summary.envs?.find((e) => e.env === v)?.health || 'undeployed';
          return (
            <span className="environment-list-health">
              <StatusBadge tone={health} label={healthLabels[health]} />
              {paused.includes(v) && <StatusBadge tone="suspended" label="Paused" />}
            </span>
          );
        }}
      />
      <Table.Column
        dataIndex="name"
        title={<Translation>Components</Translation>}
        cell={(v: string) => {
          const env = summary.envs?.find((e) => e.env === v);
          return env ? `${env.healthyComponents}/${env.components}` : '-';
        }}
      />
      <Table.Column
        dataIndex="name"
        title={<Translation>Workflow</Translation>}
        cell={(v: string) => {
          const env = summary.envs?.find((e) => e.env === v);
          return env?.workflow ? <Translation>{workflowLabel(env.workflow)}</Translation> : '-';
        }}
      />
      <Table.Column
        dataIndex="targetNames"
        title={<Translation>Targets</Translation>}
        cell={(v: string[]) => (
          <span className="environment-list-targets">
            {(v || []).map((t) => (
              <Chip key={t}>{t}</Chip>
            ))}
          </span>
        )}
      />
      <Table.Column
        dataIndex="name"
        title={<Translation>Actions</Translation>}
        width={200}
        lock="right"
        cell={(v: string) => (
          <RowLinks
            links={[
              { label: 'Open', onClick: () => dispatch(routerRedux.push(`${base(v)}/status`)) },
              { label: 'Workflow', onClick: () => dispatch(routerRedux.push(`${base(v)}/workflow`)) },
              { label: 'Instances', onClick: () => dispatch(routerRedux.push(`${base(v)}/instances`)) },
            ]}
          />
        )}
      />
    </Table>
  );
};

// EnvironmentList lists the envs an application is bound to as rows of their
// health, each expanding to its targets; its name opens the env's live view.
export const EnvironmentList = (props: {
  appName: string;
  envbinding: EnvBinding[];
  applicationAllStatus: EnvironmentStatus[];
  dispatch: (action: any) => void;
}) => {
  const { appName, envbinding, applicationAllStatus, dispatch } = props;
  const [open, setOpen] = useState<Record<string, boolean>>({});
  const summary = summariseStatuses(applicationAllStatus);
  const paused = pausedEnvs(applicationAllStatus);
  const toggle = (name: string) => setOpen({ ...open, [name]: !open[name] });
  if (envbinding.length === 0) {
    return null;
  }
  return (
    <div className="row-list environment-list">
      <div className="row-list-head">
        <span />
        <span>
          <Translation>Name</Translation>
        </span>
        <span>
          <Translation>Health</Translation>
        </span>
        <span>
          <Translation>Components</Translation>
        </span>
        <span>
          <Translation>Workflow</Translation>
        </span>
        <span>
          <Translation>Targets</Translation>
        </span>
        <span />
      </div>
      {envbinding.map((binding) => {
        const env = summary.envs?.find((e) => e.env === binding.name);
        const health = env?.health || 'undeployed';
        const expanded = !!open[binding.name];
        const view = () => dispatch(routerRedux.push(`/applications/${appName}/envbinding/${binding.name}/status`));
        return (
          <div key={binding.name} className={`row-list-row ${expanded ? 'expanded' : ''}`}>
            <div className="row-list-main">
              <span className="row-list-chevron" onClick={() => toggle(binding.name)}>
                {expanded ? <AiOutlineDown /> : <AiOutlineRight />}
              </span>
              <span className="row-list-name" onClick={() => toggle(binding.name)}>
                <AiOutlineEnvironment className="row-list-icon" />
                <span>
                  <Link
                    className="row-list-title environment-list-link"
                    to={`/applications/${appName}/envbinding/${binding.name}/status`}
                    onClick={(e) => e.stopPropagation()}
                  >
                    {binding.alias || binding.name}
                  </Link>
                  {binding.alias && binding.alias !== binding.name && (
                    <span className="row-list-type">{binding.name}</span>
                  )}
                </span>
              </span>
              <span className="environment-list-health">
                <StatusBadge tone={health} label={healthLabels[health]} />
                {paused.includes(binding.name) && <StatusBadge tone="suspended" label="Paused" />}
              </span>
              <span>
                {env ? `${env.healthyComponents}/${env.components}` : <span className="row-list-muted">-</span>}
              </span>
              <span>
                {env?.workflow ? (
                  <Translation>{workflowLabel(env.workflow)}</Translation>
                ) : (
                  <span className="row-list-muted">-</span>
                )}
              </span>
              <span className="environment-list-targets">
                {(binding.targetNames || []).map((t) => (
                  <Chip key={t}>{t}</Chip>
                ))}
              </span>
              <span className="row-list-actions">
                <RowAction icon={<AiOutlineEye />} label="Open" onClick={view} />
              </span>
            </div>
            {expanded && <EnvironmentDetail binding={binding} />}
          </div>
        );
      })}
    </div>
  );
};

// EnvironmentDetail is what an env row opens to: its targets, and what it
// deploys and with which workflow.
const EnvironmentDetail = (props: { binding: EnvBinding }) => {
  const { binding } = props;
  return (
    <div className="row-list-detail">
      {binding.description && <p className="row-list-description">{binding.description}</p>}
      <div className="row-list-detail-grid">
        <div>
          <div className="row-list-detail-title">
            <Translation>Targets</Translation>
          </div>
          {(binding.targets || []).length === 0 ? (
            <span className="row-list-muted">
              <Translation>No targets</Translation>
            </span>
          ) : (
            <dl className="row-list-properties">
              {(binding.targets || []).map((t) => (
                <React.Fragment key={t.name}>
                  <dt>{t.alias || t.name}</dt>
                  <dd>
                    {[t.clusterAlias || t.cluster?.clusterName, t.cluster?.namespace].filter(Boolean).join(' / ')}
                  </dd>
                </React.Fragment>
              ))}
            </dl>
          )}
        </div>
        <div>
          <div className="row-list-detail-title">
            <Translation>Deployment</Translation>
          </div>
          <dl className="row-list-properties">
            <dt>
              <Translation>Application</Translation>
            </dt>
            <dd>{`${binding.appDeployNamespace}/${binding.appDeployName}`}</dd>
            <dt>
              <Translation>Workflow</Translation>
            </dt>
            <dd>{binding.workflow?.alias || binding.workflow?.name || '-'}</dd>
            {binding.createTime && (
              <React.Fragment>
                <dt>
                  <Translation>Bound</Translation>
                </dt>
                <dd>
                  <RelativeTime time={binding.createTime} />
                </dd>
              </React.Fragment>
            )}
          </dl>
        </div>
      </div>
    </div>
  );
};
