// The AI console's sections, as VelaUX routes them: /ai/<section>. The
// monitor is a page of its own; the rest are the workspace's hash routes.
export const aiSections = ['overview', 'jobs', 'services', 'models', 'datasets', 'monitor'] as const;

export type AISection = (typeof aiSections)[number];

// The business codes the ticket API answers with when the console cannot open.
export const notConfiguredCode = 25001;
export const projectRequiredCode = 25002;

// sectionOf is a VelaUX route's section, the workspace's overview for one it
// does not know.
export function sectionOf(param?: string): AISection {
  return (aiSections as readonly string[]).includes(param || '') ? (param as AISection) : 'overview';
}

// hashOf is the workspace hash that shows a section.
export function hashOf(section: AISection): string {
  return section === 'monitor' ? '' : `#/${section}`;
}

// sectionOfHash is the section a workspace hash is in, as #/jobs/ns/name is
// in jobs; one outside the sections gives none.
export function sectionOfHash(hash: string): AISection | undefined {
  const head = hash.replace(/^#\/?/, '').split(/[/?]/)[0];
  if (!head) {
    return 'overview';
  }
  // The monitor is a page, not a workspace hash.
  if (head === 'monitor' || !(aiSections as readonly string[]).includes(head)) {
    return undefined;
  }
  return head as AISection;
}

// ConsoleRoute is what the console says when its route changes.
export interface ConsoleRoute {
  source: 'ai-console';
  type: 'route';
  hash: string;
  title?: string;
}

// isConsoleRoute tells the console's route messages from any other a page
// may receive.
export function isConsoleRoute(data: unknown): data is ConsoleRoute {
  const d = data as ConsoleRoute;
  return !!d && typeof d === 'object' && d.source === 'ai-console' && d.type === 'route' && typeof d.hash === 'string';
}

// navigateMessage asks the console to show a hash without reloading.
export function navigateMessage(hash: string) {
  return { source: 'velaux', type: 'navigate', hash };
}
