import { Table } from '@alifd/next';
import { Link } from 'dva/router';
import React, { Fragment, useCallback, useEffect, useState } from 'react';
import { BsBoxSeam } from 'react-icons/bs';

import type { PackageBase } from '@velaux/data';
import { listPackages } from '../../api/package';
import { ListToolbar, RowLinks, SegmentedPicker, TablePager, pageOf } from '../../components/Console';
import Empty from '../../components/Empty';
import { ListTitle } from '../../components/ListTitle';
import '../../components/RowList';
import { locationService } from '../../services/LocationService';
import { locale } from '../../utils/locale';
import { Translation } from '../../components/Translation';
import i18n from '../../i18n';
import type { PackageSource } from './packages';
import { filterPackages, packageLink } from './packages';
import './index.less';

// ProviderPill says what runs a package's functions: an external provider by
// its protocol (its endpoint on hover), CueX for a package of CUE alone, or
// KubeVela for one built in.
const ProviderPill = (props: { pkg: PackageBase }) => {
  const { pkg } = props;
  if (pkg.provider) {
    return (
      <span className="provider-pill external" title={pkg.provider.endpoint}>
        {pkg.provider.protocol.toUpperCase()}
      </span>
    );
  }
  if (pkg.builtin) {
    return (
      <span className="provider-pill builtin" title={i18n.t('Runs in KubeVela').toString()}>
        KubeVela
      </span>
    );
  }
  return (
    <span className="provider-pill cuex" title={i18n.t('CUE only, no provider').toString()}>
      CueX
    </span>
  );
};

// Packages lists the CUE packages a definition can import: the Package resources
// in the cluster, then those built into KubeVela.
const Packages = () => {
  const [packages, setPackages] = useState<PackageBase[]>([]);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState('');
  const [source, setSource] = useState<PackageSource>('all');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const load = useCallback(() => {
    setLoading(true);
    listPackages()
      .then((res: any) => setPackages(res?.packages || []))
      .finally(() => setLoading(false));
  }, []);
  useEffect(() => {
    load();
  }, [load]);
  const shown = filterPackages(packages, query, source);
  // A narrower filter keeps the page in range.
  const current = Math.min(page, Math.max(1, Math.ceil(shown.length / pageSize)));
  return (
    <div className="packages">
      <ListTitle title="Packages" subTitle="The CUE packages definitions can import, and the functions they offer" />
      <div className="console-panel">
        <div className="packages-toolbar">
          <ListToolbar
            search={{ value: query, placeholder: 'Search by name or path', live: true, onSearch: setQuery }}
            filters={
              <SegmentedPicker
                value={source}
                onChange={(v) => setSource(v as PackageSource)}
                options={[
                  { value: 'all', label: 'All' },
                  { value: 'cluster', label: 'Custom' },
                  { value: 'builtin', label: 'Built-in' },
                ]}
              />
            }
            onRefresh={load}
            refreshing={loading}
          />
        </div>
        {!loading && shown.length === 0 ? (
          <Empty message={<Translation>No packages</Translation>} />
        ) : (
          <Fragment>
            <Table
              locale={locale().Table}
              dataSource={pageOf(shown, current, pageSize)}
              loading={loading}
              primaryKey="path"
            >
              <Table.Column
                title={<Translation>Name</Translation>}
                dataIndex="name"
                width={260}
                cell={(v: string, i: number, p: PackageBase) => (
                  <Link className="package-name" to={packageLink(p)}>
                    <BsBoxSeam className="package-name-icon" />
                    <span>
                      <span className="package-name-title">{p.name}</span>
                      <span className="package-name-sub">
                        {p.builtin ? (
                          <span>
                            <Translation>Built-in</Translation>
                            {p.usedBy && p.usedBy.length > 0 && (
                              <span>
                                {' · '}
                                {p.usedBy.map((u) => i18n.t(u)).join(', ')}
                              </span>
                            )}
                          </span>
                        ) : (
                          p.namespace
                        )}
                      </span>
                    </span>
                  </Link>
                )}
              />
              <Table.Column
                title={<Translation>Import path</Translation>}
                dataIndex="path"
                cell={(v: string) => <code className="row-list-code">{v}</code>}
              />
              <Table.Column
                title={<Translation>Provider</Translation>}
                width={140}
                cell={(v: any, i: number, p: PackageBase) => <ProviderPill pkg={p} />}
              />
              <Table.Column title={<Translation>Functions</Translation>} dataIndex="functions" width={110} />
              <Table.Column
                title={<Translation>Actions</Translation>}
                lock="right"
                width={110}
                cell={(v: any, i: number, p: PackageBase) => (
                  <RowLinks links={[{ label: 'Details', onClick: () => locationService.push(packageLink(p)) }]} />
                )}
              />
            </Table>
            <TablePager
              total={shown.length}
              page={current}
              pageSize={pageSize}
              pageSizes={[20, 50, 100]}
              onChange={(p, size) => {
                setPage(p);
                setPageSize(size);
              }}
            />
          </Fragment>
        )}
      </div>
    </div>
  );
};

export default Packages;
