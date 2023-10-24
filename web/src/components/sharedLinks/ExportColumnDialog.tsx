import { type SharedLinkTableKey } from "@/constant";
import { Checkbox, Modal, Space, Table } from "antd";
import { forwardRef, useImperativeHandle, useState } from "react";

export interface ExportColumnDialogProps {
  columns: SharedLinkTableKey[];
}

export interface ExportColumnDialogRef {
  showModel(params: {
    title: string;
    defaultSelected?: React.Key[];
    footerButtonRender: (
      selectedColumns: React.Key[],
      includeHeader: boolean,
    ) => React.ReactNode;
  }): void;
  hideModel(): void;
}

const ExportColumnDialog = forwardRef<
  ExportColumnDialogRef,
  ExportColumnDialogProps
>(({ columns }, ref) => {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedRowKeys, setSelectedKeys] = useState<React.Key[]>(columns);
  const data: Array<{ key: React.Key }> = columns.map((v) => ({ key: v }));

  const hideModel: ExportColumnDialogRef["hideModel"] = () =>
    setIsModalOpen(false);

  const [modelParams, setModelParams] =
    useState<Parameters<ExportColumnDialogRef["showModel"]>[0]>();
  const showModel: ExportColumnDialogRef["showModel"] = (params) => {
    params.defaultSelected || setSelectedKeys(columns);
    setModelParams(params);
    setIsModalOpen(true);
  };

  useImperativeHandle(ref, () => ({
    showModel,
    hideModel,
  }));

  const [includeHeader, setIncludeHeader] = useState(true);

  return (
    <>
      <Modal
        title={modelParams?.title}
        open={isModalOpen}
        onCancel={hideModel}
        footer={
          <Space style={{ width: "100%", justifyContent: "space-between" }}>
            <Space>
              <Checkbox
                checked={includeHeader}
                onChange={() => setIncludeHeader(!includeHeader)}
              >
                Include Header
              </Checkbox>
            </Space>
            {modelParams?.footerButtonRender(selectedRowKeys, includeHeader)}
          </Space>
        }
      >
        <Table
          showHeader={false}
          pagination={{ hideOnSinglePage: true, pageSize: 100 }}
          size="small"
          rowSelection={{
            selectedRowKeys,
            type: "checkbox",
            onChange: setSelectedKeys,
          }}
          columns={[{ dataIndex: "key" }]}
          dataSource={data}
        />
      </Modal>
    </>
  );
});

export default ExportColumnDialog;
