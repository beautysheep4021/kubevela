import { expect } from 'chai';

import { hashOf, isConsoleRoute, navigateMessage, sectionOf, sectionOfHash } from '../../pages/AIConsole/route';

describe('AI console routes', () => {
  it('reads a VelaUX route as a section, the overview for one it does not know', () => {
    expect(sectionOf('jobs')).to.equal('jobs');
    expect(sectionOf('monitor')).to.equal('monitor');
    expect(sectionOf('nope')).to.equal('overview');
    expect(sectionOf(undefined)).to.equal('overview');
  });

  it('gives each workspace section its hash, and the monitor none, it being a page', () => {
    expect(hashOf('datasets')).to.equal('#/datasets');
    expect(hashOf('monitor')).to.equal('');
  });

  it('reads the section a console hash is in, a detail page too', () => {
    expect(sectionOfHash('#/jobs')).to.equal('jobs');
    expect(sectionOfHash('#/services/ai-shop/chat')).to.equal('services');
    expect(sectionOfHash('#/models?q=x')).to.equal('models');
    expect(sectionOfHash('')).to.equal('overview');
    expect(sectionOfHash('#/')).to.equal('overview');
    expect(sectionOfHash('#/monitor')).to.equal(undefined);
    expect(sectionOfHash('#/elsewhere')).to.equal(undefined);
  });

  it("tells the console's route messages from others", () => {
    expect(isConsoleRoute({ source: 'ai-console', type: 'route', hash: '#/jobs', title: '任务' })).to.equal(true);
    expect(isConsoleRoute({ source: 'velaux', type: 'route', hash: '#/jobs' })).to.equal(false);
    expect(isConsoleRoute({ source: 'ai-console', type: 'route' })).to.equal(false);
    expect(isConsoleRoute('#/jobs')).to.equal(false);
    expect(isConsoleRoute(null)).to.equal(false);
    expect(navigateMessage('#/jobs')).to.deep.equal({ source: 'velaux', type: 'navigate', hash: '#/jobs' });
  });
});
