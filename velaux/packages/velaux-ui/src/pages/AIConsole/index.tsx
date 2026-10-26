import { Loading } from '@alifd/next';
import { connect } from 'dva';
import React, { useCallback, useEffect, useRef, useState } from 'react';
import { TbBrain, TbPlugConnected, TbServerBolt } from 'react-icons/tb';

import { createAIConsoleTicket } from '../../api/aiConsole';
import { PageGuide, Panel } from '../../components/Console';
import Empty from '../../components/Empty';
import { Translation } from '../../components/Translation';
import i18n from '../../i18n';
import { allProjects } from '../../utils/currentProject';
import type { AISection } from './route';
import {
  hashOf,
  isConsoleRoute,
  navigateMessage,
  notConfiguredCode,
  projectRequiredCode,
  sectionOf,
  sectionOfHash,
} from './route';
import './index.less';

type Props = {
  match: { params: { section?: string } };
  history: { replace: (path: string) => void };
  currentProject?: { current: string; resolved: boolean };
};

// What the page shows instead of the console when it cannot open.
type Blocked = 'pick-project' | 'not-configured' | 'failed';

// AIConsole opens the AI console, served by the northbound server, inside
// VelaUX: a ticket for the picked project becomes the iframe's session, and
// moving between the AI pages moves the console without reloading it.
const AIConsole = (props: Props) => {
  const section = sectionOf(props.match.params.section);
  const monitor = section === 'monitor';
  const resolved = !!props.currentProject?.resolved;
  const project = props.currentProject?.current || allProjects;
  // scope is what a ticket is for; the monitor spans every project.
  const scope = monitor ? 'monitor' : project;
  const [src, setSrc] = useState<string>();
  const [blocked, setBlocked] = useState<Blocked>();
  const [message, setMessage] = useState<string>();
  const frame = useRef<HTMLIFrameElement>(null);
  // shown is the section the console shows, so a route it reports is not
  // sent back to it.
  const shown = useRef<AISection>(section);

  // A new ticket opens the console afresh: for another project, or between
  // the workspace and the monitor.
  useEffect(() => {
    if (!resolved) {
      return;
    }
    setSrc(undefined);
    setBlocked(undefined);
    if (!monitor && project === allProjects) {
      setBlocked('pick-project');
      return;
    }
    let cancelled = false;
    createAIConsoleTicket(monitor ? { monitor: true } : { project })
      .then((ticket) => {
        if (!cancelled && ticket && ticket.url) {
          shown.current = section;
          setSrc(ticket.url + hashOf(section));
        }
      })
      .catch((err: any) => {
        if (cancelled) {
          return;
        }
        const code = err && err.BusinessCode;
        setMessage(err && err.Message);
        setBlocked(
          code === notConfiguredCode ? 'not-configured' : code === projectRequiredCode ? 'pick-project' : 'failed'
        );
      });
    return () => {
      cancelled = true;
    };
    // The section moves an open workspace instead of reopening it.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [resolved, scope]);

  // Another workspace section moves the open console there.
  useEffect(() => {
    if (monitor || !src || shown.current === section) {
      return;
    }
    shown.current = section;
    frame.current?.contentWindow?.postMessage(navigateMessage(hashOf(section)), window.location.origin);
  }, [section, monitor, src]);

  // The console's own navigation keeps VelaUX's route in step.
  const onMessage = useCallback(
    (event: MessageEvent) => {
      if (event.origin !== window.location.origin || event.source !== frame.current?.contentWindow) {
        return;
      }
      if (!isConsoleRoute(event.data) || monitor) {
        return;
      }
      const next = sectionOfHash(event.data.hash);
      if (next && next !== shown.current) {
        shown.current = next;
        props.history.replace(`/ai/${next}`);
      }
    },
    [monitor, props.history]
  );
  useEffect(() => {
    window.addEventListener('message', onMessage);
    return () => window.removeEventListener('message', onMessage);
  }, [onMessage]);

  if (blocked === 'not-configured') {
    return <NotConnected />;
  }
  if (blocked === 'pick-project') {
    return (
      <Empty
        style={{ minHeight: '400px' }}
        message={<Translation>Pick a project in the top bar to open its AI workspace.</Translation>}
      />
    );
  }
  if (blocked === 'failed') {
    return (
      <Empty
        style={{ minHeight: '400px' }}
        message={message || <Translation>The AI console could not be opened.</Translation>}
      />
    );
  }
  return (
    <div className="ai-console-page">
      <Loading visible={!src} className="ai-console-loading">
        {src && (
          <iframe
            ref={frame}
            className="ai-console-frame"
            src={src}
            title={i18n.t(monitor ? 'AI Monitor' : 'AI Workspace').toString()}
          />
        )}
      </Loading>
    </div>
  );
};

// NotConnected says the AI platform is not connected, and how to connect it.
const NotConnected = () => (
  <div>
    <PageGuide
      id="ai-console"
      title="AI Platform"
      description="Train, evaluate and serve models in a project's own AI workspace."
      steps={[
        {
          icon: <TbServerBolt />,
          title: 'Deploy the AI console',
          description: <Translation>Run the northbound server, which serves the AI console and its API</Translation>,
        },
        {
          icon: <TbPlugConnected />,
          title: 'Connect it to VelaUX',
          description: <Translation>Point VelaUX at it with --ai-console-endpoint</Translation>,
        },
        {
          icon: <TbBrain />,
          title: 'Open it here',
          description: <Translation>Each project opens its own workspace, scoped to its namespace</Translation>,
        },
      ]}
    />
    <Panel title="The AI platform is not connected" description="Set the endpoint on the VelaUX server, then reload.">
      <pre className="ai-console-snippet">
        {`velaux --ai-console-endpoint=http://kubevela-ai-northbound.vela-system:8080
# or
VELAUX_AI_CONSOLE_ENDPOINT=http://kubevela-ai-northbound.vela-system:8080

# optional: the namespace each project's workspace is scoped to
velaux --ai-namespace-template='ai-{project}'`}
      </pre>
    </Panel>
  </div>
);

export default connect((store: any) => ({ currentProject: store.currentProject }))(AIConsole);
