import { Field, Form, Input, Message, Select, Switch } from '@alifd/next';
import { connect } from 'dva';
import React, { useEffect, useState } from 'react';

import type { Env, LoginUserInfo } from '@velaux/data';
import { createApplication } from '../../api/application';
import { getEnvs } from '../../api/env';
import { FormGrid, FormPage, FormSection, KVList, LabelTip, SegmentedPicker, toRecord } from '../../components/Console';
import type { KV } from '../../components/Console';
import i18n from '../../i18n';
import { locationService } from '../../services/LocationService';
import { checkName } from '../../utils/common';
import { checkPermission } from '../../utils/permission';
import { defaultEnvs, newServiceRequest } from '../ApplicationList/components/NewServiceDialog/request';
import type { NewServiceValues } from '../ApplicationList/components/NewServiceDialog/request';

type Props = {
  userInfo?: LoginUserInfo;
  currentProject?: { current: string; resolved: boolean };
  history: any;
};

// NewApplication creates an application on a page of its own, as the cloud
// consoles create a resource: who and where it is, how it is kept, and its
// metadata, in sections. Its components are added on its own page, which
// opens once it is created.
const NewApplication = (props: Props) => {
  const query = locationService.getSearchObject();
  // A project's own page fixes the project it creates in.
  const fixedProject = typeof query.project === 'string' ? query.project : undefined;
  const back = fixedProject ? `/projects/${fixedProject}/applications` : '/applications';
  const projects = (props.userInfo?.projects || []).filter((p) =>
    checkPermission({ resource: `project:${p.name}/application:*`, action: 'create' }, p.name, props.userInfo)
  );
  const picked = props.currentProject?.current;
  const initialProject =
    fixedProject || (picked && projects.find((p) => p.name === picked) ? picked : projects[0]?.name);

  const field = Field.useField({
    values: {
      project: initialProject,
      envs: [],
      expressions: false,
      paused: false,
      resyncInterval: '',
      workflowMode: 'StepByStep',
    },
  });
  const [labels, setLabels] = useState<KV[]>([]);
  const [annotations, setAnnotations] = useState<KV[]>([]);
  const [envs, setEnvs] = useState<Env[]>([]);
  const [saving, setSaving] = useState(false);
  const project = field.getValue('project') as string;

  // The user's projects arrive after the page opens; the first they may
  // create in is the default.
  useEffect(() => {
    if (!field.getValue('project') && initialProject) {
      field.setValue('project', initialProject);
    }
  }, [initialProject, field]);

  useEffect(() => {
    if (!project) {
      setEnvs([]);
      return;
    }
    getEnvs({ project }).then((res: any) => {
      const loaded: Env[] = res?.envs || [];
      setEnvs(loaded);
      field.setValue(
        'envs',
        defaultEnvs(
          loaded.map((e) => e.name),
          (field.getValue('envs') as string[]) || []
        )
      );
    });
  }, [project, field]);

  const submit = () => {
    field.validate((errors: any, values: any) => {
      if (errors) {
        return;
      }
      setSaving(true);
      const request: NewServiceValues = {
        ...(values as NewServiceValues),
        labels: toRecord(labels),
        annotations: toRecord(annotations),
      };
      createApplication(newServiceRequest(request))
        .then((res: any) => {
          if (res) {
            Message.success(i18n.t('Application created').toString());
            props.history.push(`/applications/${values.name}/config`);
          }
        })
        .finally(() => setSaving(false));
    });
  };

  const { init } = field;
  return (
    <FormPage
      title="New Application"
      crumbs={[{ label: 'Applications', to: back }, { label: 'New Application' }]}
      onBack={() => props.history.push(back)}
      onCancel={() => props.history.push(back)}
      onSubmit={submit}
      submitting={saving}
      submitLabel="Create"
    >
      <Form field={field} labelAlign="top" fullWidth>
        <FormSection title="Basic information">
          <FormGrid>
            <Form.Item label={<LabelTip label="Name" tip="Lower case letters, digits and hyphens" />} required>
              <Input
                {...init('name', {
                  rules: [
                    { required: true, message: i18n.t('Give the application a name').toString() },
                    { pattern: checkName, message: i18n.t('Lower case letters, digits and hyphens').toString() },
                  ],
                })}
                maxLength={32}
                placeholder={i18n.t('Please enter the application name').toString()}
              />
            </Form.Item>
            <Form.Item label={<LabelTip label="Alias" tip="Give your app a more recognizable name" />}>
              <Input {...init('alias')} maxLength={64} placeholder={i18n.t('Please enter').toString()} />
            </Form.Item>
            <Form.Item label={<LabelTip label="Project" tip="The project the application belongs to" />} required>
              <Select
                {...init('project', { rules: [{ required: true, message: i18n.t('Choose a project').toString() }] })}
                disabled={!!fixedProject}
                showSearch
                dataSource={projects.map((p) => ({ value: p.name, label: p.alias || p.name }))}
                placeholder={i18n.t('Please select a project').toString()}
                onChange={(v: string) => {
                  field.setValue('project', v);
                  field.setValue('envs', []);
                }}
              />
            </Form.Item>
            <Form.Item
              className="form-grid-span"
              label={<LabelTip label="Environment" tip="Where it deploys; more can be added later" />}
            >
              <Select
                {...init('envs')}
                mode="multiple"
                dataSource={envs.map((e) => ({ value: e.name, label: e.alias || e.name }))}
                placeholder={i18n.t('Choose').toString()}
                notFoundContent={i18n.t('This project has no environments yet').toString()}
              />
            </Form.Item>
            <Form.Item className="form-grid-span" label={<LabelTip label="Description" />}>
              <Input.TextArea
                {...init('description')}
                maxLength={256}
                rows={2}
                placeholder={i18n.t('Please enter').toString()}
              />
            </Form.Item>
          </FormGrid>
        </FormSection>

        <FormSection title="Delivery settings">
          <FormGrid>
            <Form.Item label={<LabelTip label="Workflow mode" tip="Run the steps one after another, or as a DAG" />}>
              <SegmentedPicker
                {...init('workflowMode')}
                options={[
                  { value: 'StepByStep', label: 'Step by step' },
                  { value: 'DAG', label: 'DAG' },
                ]}
              />
            </Form.Item>
            <Form.Item
              label={
                <LabelTip
                  label="Resync interval"
                  tip="How often the controller reconciles it, such as 5m; the default when empty"
                />
              }
            >
              <Input
                {...init('resyncInterval', {
                  rules: [
                    {
                      pattern: /^\s*(([0-9]+(\.[0-9]+)?(h|m|s))+)?\s*$/,
                      message: i18n.t('A duration such as 10m or 1h').toString(),
                    },
                  ],
                })}
                placeholder="5m"
              />
            </Form.Item>
            <div />
            <Form.Item
              label={
                <LabelTip
                  label="Expressions"
                  tip="Let properties read sources, context and other components with $( )"
                />
              }
            >
              <Switch {...init('expressions', { valueName: 'checked' })} />
            </Form.Item>
            <Form.Item
              label={
                <LabelTip
                  label="Start paused"
                  tip="Deploy it, but leave the controller from reconciling it until resumed"
                />
              }
            >
              <Switch {...init('paused', { valueName: 'checked' })} />
            </Form.Item>
          </FormGrid>
        </FormSection>

        <FormSection title="Labels and annotations" description="Optional metadata on the application">
          <FormGrid columns={2}>
            <Form.Item label={<LabelTip label="Labels" />}>
              <KVList value={labels} onChange={setLabels} addLabel="Add label" />
            </Form.Item>
            <Form.Item label={<LabelTip label="Annotations" />}>
              <KVList value={annotations} onChange={setAnnotations} addLabel="Add annotation" />
            </Form.Item>
          </FormGrid>
        </FormSection>
      </Form>
    </FormPage>
  );
};

export default connect((store: any) => ({ ...store.user, currentProject: store.currentProject }))(NewApplication);
