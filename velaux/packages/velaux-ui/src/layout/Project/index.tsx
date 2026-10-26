import { connect } from 'dva';
import { Link } from 'dva/router';
import React, { Component, Fragment } from 'react';

import { Translation } from '../../components/Translation';
import { Breadcrumb } from '../../components/Breadcrumb';
import './index.less';
import type { ProjectDetail, LoginUserInfo } from '@velaux/data';
import { checkPermission } from '../../utils/permission';
import { momentDate } from '../../utils/common';

type Props = {
  activeId: string;
  match: {
    params: {
      projectName: string;
    };
  };
  dispatch: ({}) => {};
  projectDetails?: ProjectDetail;
  location: any;
  userInfo?: LoginUserInfo;
};

@connect((store: any) => {
  return { ...store.project, ...store.user };
})
class ProjectLayout extends Component<Props> {
  componentDidMount() {
    const { params = { projectName: '' } } = this.props.match;
    this.loadProjectDetail(params.projectName);
  }

  shouldComponentUpdate(nextProps: Props) {
    const change = nextProps.match.params.projectName !== this.props.match.params.projectName;
    if (change) {
      const { params = { projectName: '' } } = nextProps.match;
      this.loadProjectDetail(params.projectName);
    }
    return true;
  }

  loadProjectDetail = async (projectName: string) => {
    this.props.dispatch({
      type: 'project/getProjectDetails',
      payload: { projectName: projectName },
    });
  };

  getNavList = () => {
    const { params = { projectName: '' } } = this.props.match;
    const { projectName } = params;
    const { userInfo } = this.props;
    const list = [
      {
        id: 'summary',
        name: <Translation>Summary</Translation>,
        to: `/projects/${projectName}/summary`,
      },
      {
        id: 'applications',
        name: <Translation>Applications</Translation>,
        to: `/projects/${projectName}/applications`,
      },
      {
        id: 'pipelines',
        name: <Translation>Pipelines</Translation>,
        to: `/projects/${projectName}/pipelines`,
      },
    ];

    if (checkPermission({ resource: `project:${projectName}/role:*`, action: 'list' }, projectName, userInfo)) {
      list.push({
        id: 'roles',
        name: <Translation>Roles</Translation>,
        to: `/projects/${projectName}/roles`,
      });
    }
    if (checkPermission({ resource: `project:${projectName}/projectUser:*`, action: 'list' }, projectName, userInfo)) {
      list.push({
        id: 'members',
        name: <Translation>Members</Translation>,
        to: `/projects/${projectName}/members`,
      });
    }

    const nav = list.map((item) => {
      const active = this.props.activeId === item.id;
      return (
        <Link
          key={item.id}
          role="tab"
          aria-selected={active}
          className={active ? 'project-tab active' : 'project-tab'}
          to={item.to}
        >
          {item.name}
        </Link>
      );
    });
    return nav;
  };

  render() {
    const menu = this.getNavList();
    const { projectDetails } = this.props;
    const projectName = projectDetails?.alias || projectDetails?.name || '';
    return (
      <Fragment>
        <div className="detail-card">
          <div className="detail-head">
            <Breadcrumb
              items={[
                {
                  to: '/projects',
                  title: <Translation>Projects</Translation>,
                },
                {
                  title: projectName,
                },
              ]}
            />
            <div className="detail-head-main">
              <h1 className="detail-head-name">{projectName}</h1>
              {projectDetails?.alias && projectDetails.alias !== projectDetails.name && (
                <span className="detail-head-id">{projectDetails.name}</span>
              )}
            </div>
            <dl className="detail-head-facts">
              <div>
                <dt>
                  <Translation>Owner</Translation>
                </dt>
                <dd>{projectDetails?.owner?.alias || projectDetails?.owner?.name || '-'}</dd>
              </div>
              <div>
                <dt>
                  <Translation>Created</Translation>
                </dt>
                <dd>{momentDate(projectDetails?.createTime) || '-'}</dd>
              </div>
              <div>
                <dt>
                  <Translation>Updated</Translation>
                </dt>
                <dd>{momentDate(projectDetails?.updateTime) || '-'}</dd>
              </div>
              {projectDetails?.description && (
                <div className="detail-head-wide">
                  <dt>
                    <Translation>Description</Translation>
                  </dt>
                  <dd title={projectDetails.description}>{projectDetails.description}</dd>
                </div>
              )}
            </dl>
          </div>
          <nav className="project-tabs" role="tablist">
            {menu}
          </nav>
        </div>
        {this.props.children}
      </Fragment>
    );
  }
}

export default ProjectLayout;
