// A tab is a page the user opened: the menu it belongs to and, deeper in,
// the resource it shows.
export type Tab = {
  key: string;
  path: string;
  label: string;
  detail?: string;
};

// Pages outside the console's frame open no tab.
const skipped = ['/', '/login', '/callback'];

// tabFor names the page at a path from the menu that holds it, or gives none
// for a page no menu holds.
export function tabFor(path: string, menu?: { to?: string; label: string }): Tab | undefined {
  if (skipped.includes(path) || !menu || !menu.to) {
    return undefined;
  }
  // The segment after the menu's own route names the resource, as in
  // /applications/<name>/config; the create page names none.
  const rest = path.slice(menu.to.length).split('/').filter(Boolean);
  const detail = rest[0] && rest[0] !== 'new' ? decodeURIComponent(rest[0]) : undefined;
  return {
    key: detail ? `${menu.to}/${rest[0]}` : menu.to,
    path,
    label: menu.label,
    detail,
  };
}

// openTab adds a tab, or moves an open one to its latest path in place; past
// max, the oldest tabs go.
export function openTab(tabs: Tab[], tab: Tab, max: number): Tab[] {
  const existing = tabs.findIndex((t) => t.key === tab.key);
  if (existing >= 0) {
    return tabs.map((t, i) => (i === existing ? tab : t));
  }
  return [...tabs, tab].slice(-max);
}

// closeTab drops a tab, and says where to go when it was the open one: its
// right neighbour, else its left, else the console's home, as a browser does.
export function closeTab(tabs: Tab[], key: string, current?: string): { tabs: Tab[]; goTo?: string } {
  const index = tabs.findIndex((t) => t.key === key);
  const next = tabs.filter((t) => t.key !== key);
  if (key !== current) {
    return { tabs: next };
  }
  const neighbour = next[Math.min(index, next.length - 1)];
  return { tabs: next, goTo: neighbour ? neighbour.path : '/' };
}
