import { Button, Select, Input } from '@alifd/next';
import React from 'react';

import './index.less';
import type { ShowMode } from '../..';
import type { Env } from '@velaux/data';
import { locale } from '../../../../utils/locale';
import i18n from '../../../../i18n';
import { AiOutlineAppstore, AiOutlineReload, AiOutlineSearch, AiOutlineUnorderedList } from 'react-icons/ai';

type Props = {
  dispatch: ({}) => {};
  envs?: Env[];
  appLabels?: string[];
  labelValue?: string[];
  setLabelValue: (labels: string[]) => void;
  getApplications: (params: any) => void;
  setMode: (mode: ShowMode) => void;
  showMode: ShowMode;
  // actions sit before refresh and the view, the page's create action.
  actions?: React.ReactNode;
};

type State = {
  targetValue: string;
  inputValue: string;
  envValue: string;
  labelValue: string[];
};

class SelectSearch extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      targetValue: '',
      envValue: '',
      inputValue: '',
      labelValue: [],
    };
    this.onChangeTarget = this.onChangeTarget.bind(this);
    this.handleChangName = this.handleChangName.bind(this);
    this.handleChangeLabel = this.handleChangeLabel.bind(this);
  }

  onChangeTarget(e: string) {
    this.setState(
      {
        targetValue: e,
      },
      () => {
        this.getApplications();
      }
    );
  }

  handleChangName(e: string) {
    this.setState({
      inputValue: e,
    });
  }

  handleChangeLabel(value: string[]) {
    const { setLabelValue } = this.props;
    let label = value ? value : [];
    setLabelValue(label);
    this.setState(
      {
        labelValue: label,
      },
      () => {
        this.getApplications();
      }
    );
  }

  onChangeEnv = (e: string) => {
    this.setState(
      {
        envValue: e,
      },
      () => {
        this.getApplications();
      }
    );
  };

  handleClickSearch = () => {
    this.getApplications();
  };

  getApplications = async () => {
    const { inputValue, envValue, labelValue } = this.state;
    const labelSelector = labelValue.join(',');
    const params = {
      query: inputValue,
      env: envValue,
      labels: labelSelector,
    };
    this.props.getApplications(params);
  };

  render() {
    const { appLabels, envs, showMode, labelValue } = this.props;
    const { inputValue, envValue } = this.state;

    const appPlaceholder = i18n.t('Search by name or description').toString();
    const labelSource = appLabels?.map((item) => {
      return {
        label: item,
        value: item,
      };
    });

    const envSource = envs?.map((env) => {
      return {
        label: env.alias || env.name,
        value: env.name,
      };
    });
    return (
      <div className="app-filter-bar">
        <Input
          innerBefore={<AiOutlineSearch onClick={this.handleClickSearch} className="app-filter-bar-search-icon" />}
          hasClear
          placeholder={appPlaceholder}
          onChange={this.handleChangName}
          onPressEnter={this.handleClickSearch}
          value={inputValue}
          className="app-filter-bar-search"
        />
        <Select
          locale={locale().Select}
          mode="single"
          label={i18n.t('Environment').toString()}
          placeholder={i18n.t('All').toString()}
          onChange={this.onChangeEnv}
          dataSource={envSource}
          className="app-filter-bar-select"
          hasClear
          value={envValue}
        />
        <Select
          hasClear
          label={i18n.t('Labels').toString()}
          placeholder={i18n.t('All').toString()}
          onChange={this.handleChangeLabel}
          showSearch
          mode="multiple"
          value={labelValue}
          className="app-filter-bar-select labels"
          dataSource={labelSource}
        />
        <div className="app-filter-bar-actions">
          {this.props.actions}
          <Button
            className="list-toolbar-icon"
            title={i18n.t('Refresh').toString()}
            aria-label={i18n.t('Refresh').toString()}
            onClick={() => this.getApplications()}
          >
            <AiOutlineReload />
          </Button>
          <div className="list-toolbar-view" role="group" aria-label={i18n.t('View').toString()}>
            <span
              className={showMode == 'table' ? 'active' : ''}
              title={i18n.t('Table').toString()}
              aria-pressed={showMode == 'table'}
              onClick={() => this.props.setMode('table')}
            >
              <AiOutlineUnorderedList />
            </span>
            <span
              className={showMode == 'card' ? 'active' : ''}
              title={i18n.t('Card').toString()}
              aria-pressed={showMode == 'card'}
              onClick={() => this.props.setMode('card')}
            >
              <AiOutlineAppstore />
            </span>
          </div>
        </div>
      </div>
    );
  }
}

export default SelectSearch;
