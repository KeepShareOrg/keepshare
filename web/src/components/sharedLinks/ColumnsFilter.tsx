import { SharedLinkTableKey, supportTableColumns } from "@/constant";
import useStore from "@/store";
import { DownOutlined } from "@ant-design/icons";
import {
  Typography,
  Button,
  Divider,
  Space,
  Checkbox,
  theme,
  Popover,
} from "antd";
import { useState } from "react";

const { Text } = Typography;

const ColumnsFilter = () => {
  const { token } = theme.useToken();

  const [visibleTableColumns, setVisibleTableColumns] = useStore((state) => [
    state.visibleTableColumns,
    state.setVisibleTableColumns,
  ]);

  const [selectedColumns, setSelectedColumns] = useState(
    Object.fromEntries(
      supportTableColumns.map((v) => [v, visibleTableColumns.includes(v)]),
    ),
  );

  const syncStoreColumns = () => {
    setSelectedColumns(
      Object.fromEntries(
        supportTableColumns.map((v) => [v, visibleTableColumns.includes(v)]),
      ),
    );
  };

  const handleChecked = (v: SharedLinkTableKey) => {
    const temp = JSON.parse(JSON.stringify(selectedColumns));
    temp[v] = !temp[v];
    setSelectedColumns(temp);
  };

  const [visible, setVisible] = useState(false);
  const handleOpenChange = (visible: boolean) => {
    visible || syncStoreColumns();
    setVisible(visible);
  };

  const handleCancel = () => {
    syncStoreColumns();
    setVisible(false);
  };

  const handleConfirm = () => {
    const checkedColumns = Object.entries(selectedColumns)
      .filter(([, v]) => v)
      .map(([k]) => k) as SharedLinkTableKey[];

    setVisibleTableColumns(checkedColumns);
    setVisible(false);
  };

  return (
    <Popover
      content={
        <Space direction="vertical" style={{ marginTop: token.marginXXS }}>
          {supportTableColumns.map((v) => {
            return (
              <Space key={v}>
                <Checkbox
                  checked={selectedColumns[v]}
                  onChange={() => handleChecked(v)}
                >
                  {v}
                </Checkbox>
              </Space>
            );
          })}
          <Divider style={{ marginBlock: token.marginXS }} />
          <Space>
            <Button onClick={handleCancel}>Cancel</Button>
            <Button type="primary" onClick={handleConfirm}>
              Confirm
            </Button>
          </Space>
        </Space>
      }
      open={visible}
      title="Select columns in the table"
      trigger="click"
      placement="bottomRight"
      arrow={false}
      onOpenChange={handleOpenChange}
    >
      <Button>
        <Text>Columns</Text>
        <DownOutlined />
      </Button>
    </Popover>
  );
};

export default ColumnsFilter;
