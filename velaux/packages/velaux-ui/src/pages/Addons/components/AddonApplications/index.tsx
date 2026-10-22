import { Link } from 'dva/router';
import { Table } from '@alifd/next';
import React, { useEffect, useState } from 'react';
import { AiOutlineAppstore } from 'react-icons/ai';

import type { ApplicationBase } from '@velaux/data';
import { getApplicationList } from '../../../../api/application';
import { RowLinks, TablePager, pageOf } from '../../../../components/Console';
import Empty from '../../../../components/Empty';
import { RelativeTime } from '../../../../components/RelativeTime';
import { StatusBadge } from '../../../../components/StatusBadge';
import { Translation } from '../../../../components/Translation';
import { locationService } from '../../../../services/LocationService';
import { locale } from '../../../../utils/locale';
import { healthLabels, healthOf } from '../../../ApplicationList/components/AppStatus/health';
import './index.less';

// addonLabel is the label naming the addon a synced application installs.
const addonLabel = 'ux.oam.dev/addon';

// AddonApplications lists the applications the enabled addons install, which
// the Services list leaves out.
export const AddonApplications = () => {
  const [apps, setApps] = useState<ApplicationBase[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  useEffect(() => {
    getApplicationList({ addons: 'only', withStatus: true })
      .then((res: any) => setApps(res?.applications || []))
      .finally(() => setLoading(false));
  }, []);
  if (!loading && apps.length === 0) {
    return <Empty message={<Translation>No addon applications</Translation>} />;
  }
  return (
    <div className="addon-app-list">
      <Table locale={locale().Table} dataSource={pageOf(apps, page, pageSize)} loading={loading} primaryKey="name">
        <Table.Column
          title={<Translation>Name</Translation>}
          dataIndex="name"
          cell={(v: string, i: number, app: ApplicationBase) => (
            <Link className="addon-app-name" to={`/applications/${app.name}/config`}>
              <AiOutlineAppstore />
              {app.alias || app.name}
            </Link>
          )}
        />
        <Table.Column
          title={<Translation>Addon</Translation>}
          dataIndex="labels"
          cell={(labels?: Record<string, string>) => labels?.[addonLabel] || '-'}
        />
        <Table.Column
          title={<Translation>Status</Translation>}
          dataIndex="status"
          cell={(v: any, i: number, app: ApplicationBase) => {
            const health = healthOf(app.status);
            return <StatusBadge tone={health} label={healthLabels[health]} />;
          }}
        />
        <Table.Column
          title={<Translation>Components</Translation>}
          dataIndex="status"
          cell={(v: any, i: number, app: ApplicationBase) =>
            app.status ? `${app.status.healthyComponents}/${app.status.components}` : '-'
          }
        />
        <Table.Column
          title={<Translation>Updated</Translation>}
          dataIndex="updateTime"
          cell={(v: string) => <RelativeTime time={v} />}
        />
        <Table.Column
          title={<Translation>Actions</Translation>}
          lock="right"
          width={120}
          cell={(v: any, i: number, app: ApplicationBase) => (
            <RowLinks
              links={[{ label: 'Details', onClick: () => locationService.push(`/applications/${app.name}/config`) }]}
            />
          )}
        />
      </Table>
      <TablePager
        total={apps.length}
        page={page}
        pageSize={pageSize}
        onChange={(p, size) => {
          setPage(p);
          setPageSize(size);
        }}
      />
    </div>
  );
};
