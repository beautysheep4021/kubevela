import React from 'react';
import { Balloon, Button } from '@alifd/next';
import { AiOutlineArrowLeft, AiOutlineInfoCircle } from 'react-icons/ai';
import { Link } from 'dva/router';

import { Translation } from '../Translation';

type Crumb = { label: string; to?: string };

type Props = {
  title: string;
  // crumbs lead back to the list the form adds to.
  crumbs?: Crumb[];
  onBack?: () => void;
  onCancel: () => void;
  onSubmit: () => void;
  submitting?: boolean;
  submitLabel?: string;
  children: React.ReactNode;
};

// FormPage is a create flow given the whole page rather than a dialog: the way
// back, the form's sections, and Cancel and Confirm fixed at the foot.
export const FormPage = (props: Props) => (
  <div className="form-page">
    {props.crumbs && props.crumbs.length > 0 && (
      <div className="form-page-crumbs">
        {props.crumbs.map((c, i) => (
          <React.Fragment key={c.label}>
            {i > 0 && <span className="form-page-crumb-sep">/</span>}
            {c.to ? (
              <Link to={c.to}>
                <Translation>{c.label}</Translation>
              </Link>
            ) : (
              <span>
                <Translation>{c.label}</Translation>
              </span>
            )}
          </React.Fragment>
        ))}
      </div>
    )}
    <div className="form-page-head">
      {props.onBack && (
        <span className="form-page-back" onClick={props.onBack} role="button" aria-label="Back">
          <AiOutlineArrowLeft />
        </span>
      )}
      <h1>
        <Translation>{props.title}</Translation>
      </h1>
    </div>
    <div className="form-page-body">{props.children}</div>
    <div className="form-page-footer">
      <Button onClick={props.onCancel}>
        <Translation>Cancel</Translation>
      </Button>
      <Button type="primary" loading={props.submitting} onClick={props.onSubmit}>
        <Translation>{props.submitLabel || 'Confirm'}</Translation>
      </Button>
    </div>
  </div>
);

// FormSection is a titled part of a form page, its heading marked by a bar.
export const FormSection = (props: { title: string; description?: string; children: React.ReactNode }) => (
  <section className="form-section">
    <h2 className="form-section-title">
      <Translation>{props.title}</Translation>
      {props.description && (
        <span className="form-section-description">
          <Translation>{props.description}</Translation>
        </span>
      )}
    </h2>
    {props.children}
  </section>
);

// FormGrid lays a section's fields out in columns, two or three a row.
export const FormGrid = (props: { columns?: 2 | 3; children: React.ReactNode }) => (
  <div className={`form-grid form-grid-${props.columns || 3}`}>{props.children}</div>
);

// LabelTip is a field's label with an info mark that explains it on hover.
export const LabelTip = (props: { label: string; tip?: string }) => (
  <span className="label-tip">
    <Translation>{props.label}</Translation>
    {props.tip && (
      <Balloon.Tooltip trigger={<AiOutlineInfoCircle className="label-tip-icon" />} align="t">
        <Translation>{props.tip}</Translation>
      </Balloon.Tooltip>
    )}
  </span>
);
