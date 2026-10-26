import React from 'react';
import { Pagination } from '@alifd/next';

import i18n from '../../i18n';
import { locale } from '../../utils/locale';

type Props = {
  total: number;
  page: number;
  pageSize: number;
  onChange: (page: number, pageSize: number) => void;
  pageSizes?: number[];
};

// TablePager is a list's foot: the total, the page size, the pages, and a
// jump to a page, as the consoles lay it out.
export const TablePager = (props: Props) => {
  const { total, page, pageSize, onChange } = props;
  return (
    <div className="table-pager">
      <Pagination
        size="small"
        total={total}
        current={page}
        pageSize={pageSize}
        pageSizeList={props.pageSizes || [10, 20, 50, 100]}
        pageSizeSelector="dropdown"
        pageSizePosition="start"
        showJump
        hideOnlyOnePage={false}
        locale={locale().Pagination}
        totalRender={(n: number) => i18n.t('Total {{count}} items', { count: n }).toString()}
        onChange={(p: number) => onChange(p, pageSize)}
        onPageSizeChange={(size: number) => onChange(1, size)}
      />
    </div>
  );
};

// pageOf is the slice of items a page shows.
export function pageOf<T>(items: T[], page: number, pageSize: number): T[] {
  return items.slice((page - 1) * pageSize, page * pageSize);
}
