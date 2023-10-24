import { SharedLinkInfo, getBlacklist, removeFromBlacklist } from "@/api/link";
import { SharedLinkTableKey } from "@/constant";
import {
  Space,
  Typography,
  theme,
  Input,
  Table,
  type MenuProps,
  Dropdown,
  message,
} from "antd";
import { type ColumnsType } from "antd/es/table";
import { useEffect, useState } from "react";
import CopyOrOpenDropMenu from "../common/CopyOrOpenDropMenu";
import useStore from "@/store";
import { MinusCircleOutlined, MoreOutlined } from "@ant-design/icons";

const getActionMenuItems = (
  info: SharedLinkInfo,
  refresh: (info: SharedLinkInfo) => void,
): MenuProps["items"] => {
  return [
    {
      key: "0",
      icon: <MinusCircleOutlined />,
      label: "Remove from Blacklist",
      onClick: async () => {
        try {
          const { error } = await removeFromBlacklist([info.original_link]);
          if (error) {
            message.error(error.message);
            return;
          }
          message.success("remove from blacklist success!");
          refresh(info);
        } catch (err) {
          console.error("remove from blacklist error: ", err);
        }
      },
    },
  ];
};

const { Search } = Input;
const { Paragraph, Title } = Typography;

interface GetColumnsParams {
  channelID?: string;
  refresh: (info: SharedLinkInfo) => void;
}
const getColumns = ({
  channelID,
  refresh,
}: GetColumnsParams): ColumnsType<SharedLinkInfo> => [
  {
    title: SharedLinkTableKey.ORIGINAL_LINKS,
    key: SharedLinkTableKey.ORIGINAL_LINKS,
    ellipsis: true,
    render: ({ original_link }: SharedLinkInfo) => {
      return (
        <CopyOrOpenDropMenu state={"OK"} copy formatLink={original_link} />
      );
    },
  },
  {
    title: SharedLinkTableKey.KEEP_SHARING_LINK,
    key: SharedLinkTableKey.KEEP_SHARING_LINK,
    ellipsis: true,
    render: ({ original_link }: SharedLinkInfo) => {
      const link = `${location.origin}/${channelID}/${encodeURIComponent(
        original_link,
      )}`;
      return <CopyOrOpenDropMenu state={"OK"} open copy formatLink={link} />;
    },
  },
  {
    title: SharedLinkTableKey.ACTION,
    key: SharedLinkTableKey.ACTION,
    ellipsis: true,
    width: 80,
    align: "center",
    render: (data) => {
      const actionMenuItems = getActionMenuItems(data, refresh);
      return (
        <Dropdown menu={{ items: actionMenuItems }} trigger={["click"]}>
          <MoreOutlined />
        </Dropdown>
      );
    },
  },
];

const BlackListTable = () => {
  const { token } = theme.useToken();

  const [columns, setColumns] = useState<ColumnsType<SharedLinkInfo>>([]);
  const [{ channel_id: channelID }, isMobile] = useStore((state) => [
    state.userInfo,
    state.isMobile,
  ]);

  const [totalCount, setTotalCount] = useState(0);
  const [tableData, setTableData] = useState<SharedLinkInfo[]>([]);

  const refreshTable = (info: SharedLinkInfo) => {
    setTableData(
      tableData.filter((v) => v.original_link !== info.original_link),
    );
  };

  useEffect(
    () => setColumns(getColumns({ channelID, refresh: refreshTable })),
    [channelID, tableData],
  );

  const [pageInfo, setPageInfo] = useState<{
    pageIndex: number;
    limit: number;
  }>({
    pageIndex: 0,
    limit: 10,
  });
  const { data, error, isLoading } = getBlacklist(pageInfo);

  useEffect(() => {
    if (!error && data?.data) {
      const { list, total } = data.data;
      // eslint-disable-next-line
      list &&
        setTableData(
          list.map((v: any) => ({ ...v, key: `${v.original_link}` })),
        );
      total && setTotalCount(total);
    }
  }, [pageInfo, isLoading]);

  const handleSizeChange = (current: number, size: number) => {
    setPageInfo({ pageIndex: current - 1, limit: size });
  };

  const [beforeSearchData, setBeforeSearchData] = useState<SharedLinkInfo[]>(
    [],
  );
  const handleSearch = (value: string) => {
    if (value.trim()) {
      const target = tableData.filter((v) => v.original_link === value);
      if (target.length > 0) {
        setBeforeSearchData(tableData);
        setTableData(target);
      }
    }
    if (value.trim() === "" && beforeSearchData.length > 0) {
      setTableData(beforeSearchData);
      setBeforeSearchData([]);
    }
  };

  return (
    <Paragraph style={{ marginTop: token.marginXL }}>
      <Title level={5}>Submission Results</Title>
      <Space>
        <Search
          allowClear
          onSearch={handleSearch}
          placeholder="Search links"
          style={{ width: isMobile ? "100%" : "500px" }}
        />
      </Space>
      <Table
        columns={columns}
        loading={isLoading}
        dataSource={tableData}
        scroll={{ x: isMobile ? "100%" : "0" }}
        style={{ marginTop: token.margin }}
        pagination={{
          total: totalCount,
          showSizeChanger: true,
          onShowSizeChange: handleSizeChange,
          onChange: handleSizeChange,
        }}
      />
    </Paragraph>
  );
};

export default BlackListTable;
