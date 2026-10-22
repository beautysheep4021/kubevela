import { expect } from 'chai';

import { splitLinks } from '../../components/Console/links';
import { closeTab, openTab, tabFor } from '../../layout/RecentTabs/tabs';
import type { Tab } from '../../layout/RecentTabs/tabs';

describe('splitLinks', () => {
  type Link = { label: string; hidden?: boolean; more?: boolean };
  const links = (n: number): Link[] => Array.from({ length: n }, (_, i) => ({ label: `l${i}` }));

  it('shows every link that fits, and a lone extra in place of More', () => {
    expect(splitLinks(links(3), 3).folded).to.have.length(0);
    expect(splitLinks(links(4), 3)).to.deep.equal({ shown: links(4), folded: [] });
  });

  it('folds the rest under More past one extra', () => {
    const { shown, folded } = splitLinks(links(6), 3);
    expect(shown.map((l) => l.label)).to.deep.equal(['l0', 'l1', 'l2']);
    expect(folded.map((l) => l.label)).to.deep.equal(['l3', 'l4', 'l5']);
  });

  it('drops hidden links before counting', () => {
    const { shown, folded } = splitLinks<Link>([...links(3), { label: 'x', hidden: true }, { label: 'y' }], 3);
    expect(shown.map((l) => l.label)).to.deep.equal(['l0', 'l1', 'l2', 'y']);
    expect(folded).to.have.length(0);
  });
});

describe('splitLinks with links that always fold', () => {
  type Link = { label: string; more?: boolean };
  const l = (label: string, more?: boolean): Link => ({ label, more });

  it('folds a link marked more even when there is room', () => {
    const { shown, folded } = splitLinks([l('a'), l('b'), l('long', true)], 2);
    expect(shown.map((x) => x.label)).to.deep.equal(['a', 'b']);
    expect(folded.map((x) => x.label)).to.deep.equal(['long']);
  });

  it('folds the overflow before the marked links', () => {
    const { shown, folded } = splitLinks([l('a'), l('b'), l('c'), l('long', true)], 2);
    expect(shown.map((x) => x.label)).to.deep.equal(['a', 'b']);
    expect(folded.map((x) => x.label)).to.deep.equal(['c', 'long']);
  });
});

describe('recent tabs', () => {
  const apps = { to: '/applications', label: 'Applications' };

  it('names a list page by its menu, and a resource page by the resource too', () => {
    expect(tabFor('/applications', apps)).to.deep.equal({
      key: '/applications',
      path: '/applications',
      label: 'Applications',
      detail: undefined,
    });
    const tab = tabFor('/applications/shop/config/components', apps);
    expect(tab?.key).to.equal('/applications/shop');
    expect(tab?.detail).to.equal('shop');
  });

  it('opens no tab for the create page apart from its list, or for a page no menu holds', () => {
    expect(tabFor('/applications/new', apps)?.key).to.equal('/applications');
    expect(tabFor('/login', apps)).to.equal(undefined);
    expect(tabFor('/somewhere', undefined)).to.equal(undefined);
  });

  const tab = (key: string): Tab => ({ key, path: key, label: key });

  it('moves an open tab in place, and keeps the newest past the limit', () => {
    const tabs = [tab('/a'), tab('/b')];
    const moved = openTab(tabs, { ...tab('/a'), path: '/a/x' }, 10);
    expect(moved.map((t) => t.path)).to.deep.equal(['/a/x', '/b']);
    expect(openTab(tabs, tab('/c'), 2).map((t) => t.key)).to.deep.equal(['/b', '/c']);
  });

  it('closing the open tab goes to its neighbour, and home when none is left', () => {
    const tabs = [tab('/a'), tab('/b'), tab('/c')];
    expect(closeTab(tabs, '/b', '/b').goTo).to.equal('/c');
    expect(closeTab(tabs, '/c', '/c').goTo).to.equal('/b');
    expect(closeTab(tabs, '/a', '/c').goTo).to.equal(undefined);
    expect(closeTab([tab('/a')], '/a', '/a')).to.deep.equal({ tabs: [], goTo: '/' });
  });
});
