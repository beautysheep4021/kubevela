import React, { useState } from 'react';
import { AiOutlineDoubleRight, AiOutlineDown, AiOutlineUp } from 'react-icons/ai';

import { Translation } from '../Translation';
import { readPref, writePref } from '../../utils/prefs';

export type GuideStep = {
  icon: React.ReactNode;
  title: string;
  description: React.ReactNode;
};

type Props = {
  // id names the guide, so its folded state is remembered per page.
  id: string;
  title: string;
  description: string;
  steps: GuideStep[];
  // extra sits under the title, before the steps: a page's own links.
  extra?: React.ReactNode;
};

// PageGuide opens a page with what it is for and the steps to use it, the way
// the cloud consoles introduce a service. The steps fold away once read.
export const PageGuide = (props: Props) => {
  const key = `guide-folded-${props.id}`;
  const [folded, setFolded] = useState<boolean>(readPref<boolean>(key, false));
  const toggle = () => {
    setFolded(!folded);
    writePref(key, !folded);
  };
  return (
    <div className="page-guide">
      <div className="page-guide-head">
        <div className="page-guide-text">
          <h1 className="page-guide-title">
            <Translation>{props.title}</Translation>
          </h1>
          <p className="page-guide-description">
            <Translation>{props.description}</Translation>
          </p>
        </div>
        <span className="page-guide-toggle" onClick={toggle}>
          {folded ? <AiOutlineDown /> : <AiOutlineUp />}
          <Translation>{folded ? 'Show steps' : 'Hide steps'}</Translation>
        </span>
      </div>
      {props.extra}
      {!folded && (
        <div className="page-guide-steps">
          {props.steps.map((step, i) => (
            <React.Fragment key={step.title}>
              {i > 0 && (
                <div className="page-guide-arrow" aria-hidden="true">
                  <AiOutlineDoubleRight />
                </div>
              )}
              <div className="page-guide-step">
                <div className="page-guide-icon">{step.icon}</div>
                <div className="page-guide-step-title">
                  <span className="page-guide-step-index">{i + 1}</span>
                  <Translation>{step.title}</Translation>
                </div>
                <div className="page-guide-step-description">{step.description}</div>
              </div>
            </React.Fragment>
          ))}
        </div>
      )}
    </div>
  );
};
