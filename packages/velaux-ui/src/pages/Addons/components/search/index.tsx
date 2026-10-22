import { Checkbox, Select } from '@alifd/next';
import React from 'react';

import { ListToolbar } from '../../../../components/Console';
import { Translation } from '../../../../components/Translation';
import i18n from '../../../../i18n';
import { locale } from '../../../../utils/locale';
import './index.less';

type Props = {
  dispatch: ({}) => {};
  listFunction: ({}) => {};
  onTagChange: (tags: string[]) => void;
  tagList?: Array<{ tag: string; num: number }>;
  registries?: [];
  extButtons?: [React.ReactNode];
};

type State = {
  registryValue: string;
  inputValue: string;
};

const { Option } = Select;

// SelectSearch is the catalog's toolbar: the name search and registry filter
// at the left, refresh and the registries at the right, and the tags below.
class SelectSearch extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      registryValue: '',
      inputValue: '',
    };
    this.handleChangRegistry = this.handleChangRegistry.bind(this);
    this.handleChangName = this.handleChangName.bind(this);
  }

  handleChangRegistry(e: string) {
    this.setState(
      {
        registryValue: e,
      },
      () => {
        this.handleClickSearch();
      }
    );
  }

  handleChangName(e: string) {
    this.setState({ inputValue: e });
  }

  handleClickSearch = () => {
    const { listFunction } = this.props;
    const { registryValue, inputValue } = this.state;
    listFunction({ registry: registryValue, query: inputValue });
  };

  generateTagList = () => {
    const { tagList } = this.props;
    const data: { label: string; value: string }[] = [];
    tagList?.map((tag) => {
      if (tag.num > 2) {
        data.push({ label: tag.tag, value: tag.tag });
      }
    });
    return data;
  };

  render() {
    const { registries, extButtons } = this.props;
    const { registryValue, inputValue } = this.state;
    const tags = this.generateTagList();

    return (
      <div className="addon-search">
        <ListToolbar
          search={{
            value: inputValue,
            placeholder: 'Search by Name and Description etc',
            onSearch: (value: string) => this.setState({ inputValue: value }, this.handleClickSearch),
          }}
          filters={
            <Select
              locale={locale().Select}
              mode="single"
              label={i18n.t('Registry').toString()}
              onChange={this.handleChangRegistry}
              className="addon-search-registry"
              value={registryValue}
            >
              <Option value="">
                <Translation>All</Translation>
              </Option>
              {registries?.map((item: any) => {
                return (
                  <Option key={item.name} value={item.name}>
                    {item.name}
                  </Option>
                );
              })}
            </Select>
          }
          onRefresh={this.handleClickSearch}
          actions={extButtons}
        />
        {tags.length > 0 && (
          <div className="tag-search">
            <div className="tag-name">
              <Translation>Tags</Translation>
            </div>
            <div className="tag-list">
              <Checkbox.Group
                dataSource={tags}
                onChange={(tags) => {
                  this.props.onTagChange(tags);
                }}
              />
            </div>
          </div>
        )}
      </div>
    );
  }
}

export default SelectSearch;
