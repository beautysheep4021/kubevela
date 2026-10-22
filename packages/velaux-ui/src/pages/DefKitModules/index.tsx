import { Button, Table } from '@alifd/next';
import { Link, routerRedux } from 'dva/router';
import React, { Fragment, useCallback, useEffect, useState } from 'react';
import { connect } from 'dva';
import { BsBoxes } from 'react-icons/bs';

import type { DefKitModule } from '@velaux/data';
import { listDefKitModules } from '../../api/defkit';
import { ListToolbar, RowLinks, TablePager, pageOf } from '../../components/Console';
import Empty from '../../components/Empty';
import { ListTitle } from '../../components/ListTitle';
import Permission from '../../components/Permission';
import { RelativeTime } from '../../components/RelativeTime';
import '../../components/RowList';
import { StatusBadge } from '../../components/StatusBadge';
import { Translation } from '../../components/Translation';
import i18n from '../../i18n';
import { locationService } from '../../services/LocationService';
import { locale } from '../../utils/locale';
import '../Packages/index.less';
import { autoText, busy, kindLabels, phaseLabels, phaseTones, sourceText } from './defkit';
import { ModuleDialog } from './ModuleDialog';
import './index.less';

// countsText is a module's installed definitions by kind, as one line.
const countsText = (counts: Record<string, number>) =>
  Object.keys(kindLabels)
    .filter((k) => counts[k])
    .map((k) => `${counts[k]} ${i18n.t(kindLabels[k]).toString().toLowerCase()}`)
    .join(' · ');

// DefKitModules lists the DefKit modules installed as Applications of the
// defkit addon.
const DefKitModules = (props: { dispatch: (action: any) => void }) => {
  const [modules, setModules] = useState<DefKitModule[]>([]);
  const [addonEnabled, setAddonEnabled] = useState(true);
  const [loading, setLoading] = useState(true);
  const [adding, setAdding] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const load = useCallback(
    () =>
      listDefKitModules()
        .then((res: any) => {
          setModules(res?.modules || []);
          setAddonEnabled(!!res?.addonEnabled);
        })
        .finally(() => setLoading(false)),
    []
  );
  useEffect(() => {
    load();
  }, [load]);
  const anyBusy = modules.some((m) => busy(m.phase));
  useEffect(() => {
    if (!anyBusy) {
      return undefined;
    }
    const timer = setInterval(load, 5000);
    return () => clearInterval(timer);
  }, [anyBusy, load]);
  const current = Math.min(page, Math.max(1, Math.ceil(modules.length / pageSize)));
  return (
    <div className="packages defkit">
      <ListTitle
        title="DefKit"
        subTitle="Definition modules written in Go with DefKit, rendered and reviewed before they install"
      />
      <div className="defkit-experimental">
        <span className="defkit-experimental-pill">
          <Translation>Experimental</Translation>
        </span>
        <Translation>
          Each module installs through its own workflow in vela-system, so its definitions are tracked and garbage
          collected like any other resource.
        </Translation>
      </div>
      {!loading && !addonEnabled && (
        <div className="package-issue">
          <Translation>Enable the defkit addon to install DefKit modules</Translation>
          {' · '}
          <Link to="/addons">
            <Translation>Addons</Translation>
          </Link>
        </div>
      )}
      <div className="console-panel">
        <div className="defkit-toolbar">
          <ListToolbar
            onRefresh={load}
            refreshing={loading}
            actions={
              <Permission request={{ resource: 'defkit:*', action: 'create' }} project={''}>
                <Button type="primary" disabled={!addonEnabled} onClick={() => setAdding(true)}>
                  <Translation>Add module</Translation>
                </Button>
              </Permission>
            }
          />
        </div>
        {!loading && modules.length === 0 ? (
          <Empty message={<Translation>No DefKit modules</Translation>} />
        ) : (
          <Fragment>
            <Table
              locale={locale().Table}
              dataSource={pageOf(modules, current, pageSize)}
              loading={loading}
              primaryKey="name"
            >
              <Table.Column
                title={<Translation>Name</Translation>}
                dataIndex="name"
                width={240}
                cell={(v: string, i: number, m: DefKitModule) => (
                  <Link className="package-name" to={`/defkit/${m.name}`}>
                    <BsBoxes className="package-name-icon" />
                    <span>
                      <span className="package-name-title">{m.name}</span>
                      <span className="package-name-sub">{m.info?.description || m.info?.name}</span>
                    </span>
                  </Link>
                )}
              />
              <Table.Column
                title={<Translation>Source</Translation>}
                dataIndex="source"
                cell={(v: any, i: number, m: DefKitModule) => (
                  <code className="row-list-code defkit-source" title={sourceText(m.source)}>
                    {sourceText(m.source)}
                  </code>
                )}
              />
              <Table.Column
                title={<Translation>Status</Translation>}
                dataIndex="phase"
                width={180}
                cell={(v: any, i: number, m: DefKitModule) => (
                  <span>
                    <StatusBadge tone={phaseTones[m.phase]} label={phaseLabels[m.phase]} title={m.message} />
                    {m.settings?.autoUpdate && (
                      <span className="defkit-auto">
                        <Translation>Auto update every</Translation> {autoText(m.settings)}
                      </span>
                    )}
                  </span>
                )}
              />
              <Table.Column
                title={<Translation>Definitions</Translation>}
                dataIndex="counts"
                cell={(v: any, i: number, m: DefKitModule) => countsText(m.counts) || '-'}
              />
              <Table.Column
                title={<Translation>Updated</Translation>}
                dataIndex="updateTime"
                width={120}
                cell={(v: string) => <RelativeTime time={v} />}
              />
              <Table.Column
                title={<Translation>Actions</Translation>}
                lock="right"
                width={140}
                cell={(v: any, i: number, m: DefKitModule) => (
                  <RowLinks
                    links={[
                      { label: 'Details', onClick: () => locationService.push(`/defkit/${m.name}`) },
                      {
                        label: 'Review',
                        hidden: m.phase !== 'review',
                        onClick: () => locationService.push(`/defkit/${m.name}?tab=review`),
                      },
                    ]}
                  />
                )}
              />
            </Table>
            <TablePager
              total={modules.length}
              page={current}
              pageSize={pageSize}
              onChange={(p, size) => {
                setPage(p);
                setPageSize(size);
              }}
            />
          </Fragment>
        )}
      </div>
      {adding && (
        <ModuleDialog
          onClose={() => setAdding(false)}
          onDone={(name) => {
            setAdding(false);
            props.dispatch(routerRedux.push(`/defkit/${name}?tab=review`));
          }}
        />
      )}
    </div>
  );
};

export default connect()(DefKitModules);
