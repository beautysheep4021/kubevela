import { Balloon, Button, Dialog, Message, Table } from '@alifd/next';
import { connect } from 'dva';
import { Link, routerRedux } from 'dva/router';
import React, { useCallback, useEffect, useState } from 'react';
import { BsDiagram3 } from 'react-icons/bs';

import type { ListSharedWorkflowsResponse, LoginUserInfo, SharedWorkflow } from '@velaux/data';
import { deleteSharedWorkflow, listSharedWorkflows } from '../../api/sharedWorkflows';
import Empty from '../../components/Empty';
import { ListTitle } from '../../components/ListTitle';
import { ListToolbar, RowLinks, TablePager, pageOf } from '../../components/Console';
import '../../components/RowList';
import { StatusBadge } from '../../components/StatusBadge';
import { Translation } from '../../components/Translation';
import { SettingsSummary } from '../../components/WorkflowStudio/settings';
import i18n from '../../i18n';
import { allProjects } from '../../utils/currentProject';
import { locale } from '../../utils/locale';
import type { SharedUsage, SharedWhere } from '../../utils/sharedWorkflows';
import { filterShared, inUse } from '../../utils/sharedWorkflows';
import type { SharedWorkflowDraft } from './draft';
import { canChange, canCreate, studioPath } from './draft';
import type { NewSharedWorkflow } from './NewDialog';
import { NewDialog } from './NewDialog';
import '../Packages/index.less';
import './index.less';

type Props = {
  dispatch: (action: any) => void;
  currentProject?: { current: string; resolved: boolean };
  userInfo?: LoginUserInfo;
};

// UsedBy says what runs a shared workflow: the project's workflows by name on
// hover, other projects' as a count.
const UsedBy = (props: { workflow: SharedWorkflow }) => {
  const used = props.workflow.usedBy || [];
  const elsewhere = props.workflow.usedElsewhere || 0;
  if (used.length === 0 && elsewhere === 0) {
    return (
      <span className="row-list-muted">
        <Translation>Not used</Translation>
      </span>
    );
  }
  const text = (
    <span>
      {used.length > 0 && `${used.length} ${i18n.t(used.length === 1 ? 'workflow' : 'workflows').toString()}`}
      {used.length > 0 && elsewhere > 0 && ' · '}
      {elsewhere > 0 && `${elsewhere} ${i18n.t('in other projects').toString()}`}
    </span>
  );
  if (used.length === 0) {
    return text;
  }
  return (
    <Balloon trigger={<span className="shared-used-by">{text}</span>} closable={false} align="b">
      <ul className="shared-used-list">
        {used.map((u) => (
          <li key={`${u.appName}/${u.workflowName}`}>
            <Link to={`/applications/${u.appName}/envbinding/${u.envName}/workflow`}>{u.appAlias || u.appName}</Link>
            <span className="row-list-muted">
              {' '}
              · {u.workflowAlias || u.workflowName} · {u.envName}
            </span>
          </li>
        ))}
      </ul>
    </Balloon>
  );
};

// SharedWorkflows lists the shared workflows the picked project's
// applications can use: its own, in its namespace, and the global ones.
const SharedWorkflows = (props: Props) => {
  const project = props.currentProject?.resolved ? props.currentProject.current : undefined;
  const [list, setList] = useState<ListSharedWorkflowsResponse>();
  const [loading, setLoading] = useState(false);
  const [query, setQuery] = useState('');
  const [where, setWhere] = useState<SharedWhere>('all');
  const [usage, setUsage] = useState<SharedUsage>('all');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  // creating is the dialog's starting values, with the workflow being copied.
  const [creating, setCreating] = useState<{ initial: NewSharedWorkflow; from?: SharedWorkflow }>();

  const load = useCallback(() => {
    if (!project) {
      return;
    }
    setLoading(true);
    listSharedWorkflows(project)
      .then((res: ListSharedWorkflowsResponse) => setList(res))
      .finally(() => setLoading(false));
  }, [project]);
  useEffect(load, [load]);

  if (!project) {
    return null;
  }
  const canProject = canCreate(project, 'project', props.userInfo);
  const canGlobal = canCreate(project, 'global', props.userInfo);

  const remove = (w: SharedWorkflow) =>
    Dialog.confirm({
      type: 'confirm',
      content: (
        <span>
          <Translation>Delete the shared workflow</Translation> <code>{w.name}</code>?
        </span>
      ),
      onOk: () =>
        deleteSharedWorkflow(project, w.scope, w.name).then((res: any) => {
          if (res) {
            Message.success(i18n.t('Shared workflow deleted').toString());
            load();
          }
        }),
      locale: locale().Dialog,
    });

  const open = (values: NewSharedWorkflow, from?: SharedWorkflow) => {
    const draft: SharedWorkflowDraft = {
      ...values,
      project,
      namespace: values.scope === 'project' ? list?.projectNamespace : 'vela-system',
      description: from?.description,
      mode: from?.mode,
      subMode: from?.subMode,
      steps: from?.steps || [],
    };
    setCreating(undefined);
    props.dispatch(routerRedux.push({ pathname: '/shared-workflows/new', state: draft }));
  };

  const workflows = list?.workflows || [];
  // Each row is keyed by its scope and name, as a project and a global
  // workflow may share a name.
  const shown = filterShared(workflows, query, where, usage).map((w) => ({ ...w, key: `${w.scope}/${w.name}` }));
  // A filter that shrinks the list keeps the page in range.
  const currentPage = Math.min(page, Math.max(1, Math.ceil(shown.length / pageSize)));
  const newButton = (
    <Button
      key="new"
      type="primary"
      disabled={!canProject && !canGlobal}
      onClick={() => setCreating({ initial: { name: '', scope: canProject ? 'project' : 'global' } })}
    >
      <Translation>New Workflow</Translation>
    </Button>
  );
  // segmented is a filter as a row of buttons, one of them pressed.
  const segmented = <T extends string>(label: string, value: T, set: (v: T) => void, options: [T, string][]) => (
    <div className="packages-source" role="group" aria-label={i18n.t(label).toString()}>
      {options.map(([v, text]) => (
        <button
          key={v}
          type="button"
          className={value === v ? 'active' : ''}
          aria-pressed={value === v}
          onClick={() => set(v)}
        >
          <Translation>{text}</Translation>
        </button>
      ))}
    </div>
  );
  return (
    <div className="shared-workflows">
      <ListTitle
        title="Workflows"
        subTitle="Workflows the project's applications can run in place of their own steps"
        extButtons={project === allProjects || (!loading && workflows.length === 0) ? [newButton] : undefined}
      />
      {project === allProjects ? (
        <Empty
          message={
            <Translation>Pick a project in the top bar. Each project has shared workflows of its own.</Translation>
          }
        />
      ) : !loading && workflows.length === 0 ? (
        <Empty message={<Translation>No shared workflows for this project, nor global ones</Translation>} />
      ) : (
        <div className="console-panel">
          <div className="console-panel-body shared-toolbar">
            <ListToolbar
              search={{
                value: query,
                placeholder: 'Search by name, alias or description',
                live: true,
                onSearch: (v: string) => {
                  setQuery(v);
                  setPage(1);
                },
              }}
              filters={
                <>
                  {segmented<SharedWhere>('Where', where, setWhere, [
                    ['all', 'All'],
                    ['project', 'Project'],
                    ['global', 'Global'],
                  ])}
                  {segmented<SharedUsage>('Usage', usage, setUsage, [
                    ['all', 'All'],
                    ['used', 'In use'],
                    ['unused', 'Not used'],
                  ])}
                </>
              }
              onRefresh={load}
              refreshing={loading}
              actions={newButton}
            />
            {list?.globalUnavailable && (
              <div className="shared-notice">
                <Translation>Global shared workflows could not be loaded.</Translation>
              </div>
            )}
          </div>
          <Table
            locale={locale().Table}
            dataSource={pageOf(shown, currentPage, pageSize)}
            loading={loading}
            primaryKey="key"
            emptyContent={<Translation>No workflows match the filters</Translation>}
            rowProps={(w: SharedWorkflow) => ({ className: w.hidden ? 'shared-hidden' : '' })}
          >
            <Table.Column
              title={<Translation>Name</Translation>}
              dataIndex="name"
              cell={(v: string, i: number, w: SharedWorkflow) => (
                <Link className="shared-name" to={studioPath(project, w.scope, w.name)}>
                  <BsDiagram3 className="row-list-icon" />
                  <span>
                    <span className="row-list-title">{w.alias || w.name}</span>
                    <span className="row-list-type">
                      {w.alias ? w.name : ''}
                      {w.alias && w.description ? ' · ' : ''}
                      {w.description}
                    </span>
                  </span>
                </Link>
              )}
            />
            <Table.Column
              title={<Translation>Where</Translation>}
              dataIndex="scope"
              width={170}
              cell={(v: string, i: number, w: SharedWorkflow) => (
                <span>
                  <StatusBadge
                    tone={w.scope === 'global' ? 'neutral' : 'progressing'}
                    label={w.scope === 'global' ? 'Global' : 'Project'}
                  />
                  {w.hidden && (
                    <span className="row-list-muted shared-hidden-note">
                      <Translation>{"hidden by the project's"}</Translation>
                    </span>
                  )}
                </span>
              )}
            />
            <Table.Column
              title={<Translation>Runs</Translation>}
              dataIndex="mode"
              cell={(v: string, i: number, w: SharedWorkflow) => (
                <span>
                  <SettingsSummary mode={w.mode || 'StepByStep'} subMode={w.subMode || 'DAG'} />
                  <span className="row-list-muted">
                    {' · '}
                    {w.steps.length} {i18n.t(w.steps.length === 1 ? 'step' : 'steps').toString()}
                  </span>
                </span>
              )}
            />
            <Table.Column
              title={<Translation>Used by</Translation>}
              dataIndex="usedBy"
              cell={(v: string, i: number, w: SharedWorkflow) => <UsedBy workflow={w} />}
            />
            <Table.Column
              title={<Translation>Actions</Translation>}
              dataIndex="operation"
              width={180}
              lock="right"
              cell={(v: string, i: number, w: SharedWorkflow) => {
                const editable = canChange(project, w.scope, props.userInfo);
                return (
                  <RowLinks
                    links={[
                      {
                        label: editable ? 'Edit' : 'View',
                        onClick: () => props.dispatch(routerRedux.push(studioPath(project, w.scope, w.name))),
                      },
                      {
                        label: 'Copy',
                        hidden: !canProject && !canGlobal,
                        onClick: () =>
                          setCreating({
                            initial: {
                              name: `${w.name}-copy`,
                              alias: w.alias,
                              scope: canProject ? 'project' : 'global',
                            },
                            from: w,
                          }),
                      },
                      {
                        label: 'Delete',
                        danger: true,
                        hidden: !editable,
                        disabled: inUse(w),
                        reason: 'In use, so it cannot be deleted',
                        onClick: () => remove(w),
                      },
                    ]}
                  />
                );
              }}
            />
          </Table>
          <TablePager
            total={shown.length}
            page={currentPage}
            pageSize={pageSize}
            onChange={(p, size) => {
              setPage(p);
              setPageSize(size);
            }}
          />
        </div>
      )}
      {creating && (
        <NewDialog
          copyOf={creating.from?.name}
          initial={creating.initial}
          canProject={canProject}
          canGlobal={canGlobal}
          onClose={() => setCreating(undefined)}
          onOk={(values) => open(values, creating.from)}
        />
      )}
    </div>
  );
};

export default connect((store: any) => ({ currentProject: store.currentProject, userInfo: store.user?.userInfo }))(
  SharedWorkflows
);
