import React from 'react';

import { Translation } from '../Translation';

type Props = {
  title?: string;
  description?: string;
  actions?: React.ReactNode;
  // flush drops the body's padding, for a table edge to edge.
  flush?: boolean;
  children: React.ReactNode;
  className?: string;
};

// Panel is the console's white container: an optional header with its own
// actions, over the content.
export const Panel = (props: Props) => (
  <div className={`console-panel ${props.className || ''}`}>
    {(props.title || props.actions) && (
      <div className="console-panel-head">
        <div>
          {props.title && (
            <h2 className="console-panel-title">
              <Translation>{props.title}</Translation>
            </h2>
          )}
          {props.description && (
            <p className="console-panel-description">
              <Translation>{props.description}</Translation>
            </p>
          )}
        </div>
        {props.actions && <div className="console-panel-actions">{props.actions}</div>}
      </div>
    )}
    <div className={props.flush ? 'console-panel-body flush' : 'console-panel-body'}>{props.children}</div>
  </div>
);
