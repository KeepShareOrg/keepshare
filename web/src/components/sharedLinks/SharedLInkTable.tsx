import {
  DeleteOutlined,
  MinusCircleOutlined,
  MoreOutlined,
  PlusCircleOutlined,
} from "@ant-design/icons";
import {
  Badge,
  Dropdown,
  MenuProps,
  Table,
  Tooltip,
  message,
  theme,
} from "antd";
import type { ColumnsType } from "antd/es/table";
import type { TableRowSelection } from "antd/es/table/interface";
import React, { useEffect, useState } from "react";
import IconAutoShare from "@/assets/images/auto-share.png";
import IconLinkToShare from "@/assets/images/link-to-share.png";
import { TableCellIcon } from "./style";
import useStore from "@/store";
import {
  addToBlacklist,
  deleteSharedLink,
  removeFromBlacklist,
  type SharedLinkInfo,
} from "@/api/link";
import { CreatedBy, LinkFormatType, SharedLinkTableKey } from "@/constant";
import SharedLinkTableFooter from "./SharedLinkTableFooter";
import { formatLinkWithType } from "@/util";
import CopyOrOpenDropMenu from "../common/CopyOrOpenDropMenu";
import dayjs from "dayjs";

const getActionMenuItems = (
  info: SharedLinkInfo,
  refresh: VoidFunction,
): MenuProps["items"] => {
  const actions: MenuProps["items"] = [
    {
      key: "2",
      icon: <DeleteOutlined />,
      label: "Delete",
      onClick: async () => {
        try {
          const { error } = await deleteSharedLink({
            links: [info.original_link],
            host: info.host,
          });
          error && message.error(error.message);
          refresh();
        } catch (err) {
          console.error("delete shared link error: ", err);
        }
      },
    },
  ];

  if (info.state === "BLOCKED") {
    actions.unshift({
      key: "1",
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
          refresh();
        } catch (err) {
          console.error("remove from black list error: ", err);
        }
      },
    });
  }
  if (info.state === "OK") {
    actions.unshift({
      key: "0",
      icon: <PlusCircleOutlined />,
      label: "Add to Blacklist",
      onClick: async () => {
        try {
          const { error } = await addToBlacklist([info.original_link]);
          if (error) {
            message.error(error.message);
            return;
          }
          message.success("add to blacklist success!");
          refresh();
        } catch (err) {
          console.error("add tot blacklist error: ", err);
        }
      },
    });
  }

  return actions;
};

interface GetColumnsParams {
  formatType: LinkFormatType;
  channelID?: string;
  refresh: VoidFunction;
}
const getColumns = ({
  formatType,
  channelID,
  refresh,
}: GetColumnsParams): ColumnsType<SharedLinkInfo> => [
  {
    title: SharedLinkTableKey.ORIGINAL_LINKS,
    key: SharedLinkTableKey.ORIGINAL_LINKS,
    ellipsis: true,
    width: 165,
    render: ({ original_link }: SharedLinkInfo) => {
      const formatLink = formatLinkWithType(original_link, formatType);

      return <CopyOrOpenDropMenu state={"OK"} copy formatLink={formatLink} />;
    },
  },
  {
    title: SharedLinkTableKey.KEEP_SHARING_LINK,
    key: SharedLinkTableKey.KEEP_SHARING_LINK,
    ellipsis: true,
    width: 165,
    render: ({ state, created_by, original_link }: SharedLinkInfo) => {
      const link = `${location.origin}/${channelID}/${encodeURIComponent(
        original_link,
      )}`;
      const isAutoShared = created_by === CreatedBy.AUTO_SHARE;
      const formatLink = formatLinkWithType(link, formatType);
      if (state === "CREATED") {
        state = "OK";
      }

      return (
        <CopyOrOpenDropMenu state={state} open copy formatLink={formatLink}>
          <Tooltip title={isAutoShared ? "Auto Share" : "Shared Links"}>
            <TableCellIcon
              src={isAutoShared ? IconAutoShare : IconLinkToShare}
            />
          </Tooltip>
        </CopyOrOpenDropMenu>
      );
    },
  },
  {
    title: SharedLinkTableKey.TITLE,
    key: SharedLinkTableKey.TITLE,
    dataIndex: "title",
    ellipsis: true,
    width: 125,
  },
  {
    title: SharedLinkTableKey.HOST_SHARED_LINK,
    key: SharedLinkTableKey.HOST_SHARED_LINK,
    ellipsis: true,
    width: 155,
    render: ({ state, host_shared_link }: SharedLinkInfo) => {
      const formatLink = formatLinkWithType(host_shared_link, formatType);

      return (
        <CopyOrOpenDropMenu state={state} open copy formatLink={formatLink} />
      );
    },
  },
  {
    title: SharedLinkTableKey.CREATED_AT,
    key: SharedLinkTableKey.CREATED_AT,
    dataIndex: "created_at",
    ellipsis: true,
    width: 155,
  },
  {
    title: SharedLinkTableKey.VISITOR,
    key: SharedLinkTableKey.VISITOR,
    dataIndex: "visitor",
    width: 120,
    ellipsis: true,
  },
  {
    title: SharedLinkTableKey.STORED,
    key: SharedLinkTableKey.STORED,
    dataIndex: "stored",
    width: 120,
    ellipsis: true,
  },
  {
    title: SharedLinkTableKey.SIZE,
    key: SharedLinkTableKey.SIZE,
    dataIndex: "size",
    width: 120,
    ellipsis: true,
  },
  {
    title: SharedLinkTableKey.DAYS_NOT_VISIT,
    key: SharedLinkTableKey.DAYS_NOT_VISIT,
    width: 120,
    ellipsis: true,
    render: ({ last_visited_at }: SharedLinkInfo) => {
      return <> {dayjs().diff(dayjs(last_visited_at), "day")} </>;
    },
  },
  {
    title: SharedLinkTableKey.STATE,
    key: SharedLinkTableKey.STATE,
    ellipsis: true,
    width: 120,
    render: ({ state }: SharedLinkInfo) => {
      return (
        <>
          <Badge
            status={
              state === "OK"
                ? "success"
                : ["CREATED", "PENDING"].includes(state)
                ? "warning"
                : "error"
            }
            text={state}
          />
        </>
      );
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

interface ComponentProps {
  refresh: VoidFunction;
  formatType: LinkFormatType;
  handlePageChange: (pageIndex: number, pageSize: number) => void;
}
const SharedLinkTable = ({
  formatType,
  handlePageChange,
  refresh,
}: ComponentProps) => {
  const { token } = theme.useToken();

  const [
    isLoading,
    tableData,
    totalSharedNum,
    selectedSharedLinkKeys,
    setSelectedSharedLinkKeys,
    totalSharedLinks,
  ] = useStore((state) => [
    state.isLoading,
    state.tableData,
    state.totalSharedNum,
    state.selectedSharedLinkKeys,
    state.setSelectedSharedLinkKeys,
    state.totalSharedLinks,
  ]);

  const [selectedSharedLinks, setSelectedSharedLinks] = useState<
    SharedLinkInfo[]
  >([]);
  useEffect(() => {
    setSelectedSharedLinks(
      totalSharedLinks.filter((v) =>
        selectedSharedLinkKeys.includes(`${v.auto_id}`),
      ),
    );
  }, [selectedSharedLinkKeys]);

  const onSelectChange = (rowKeys: React.Key[]) => {
    const shouldRemoveKeys = tableData
      .map((v) => `${v.auto_id}`)
      .filter((v) => !rowKeys.includes(v));
    const shouldAddKeys = rowKeys.filter(
      (v) => !selectedSharedLinkKeys.includes(v),
    );

    const newSelectedKeys = selectedSharedLinkKeys
      .concat(shouldAddKeys)
      .filter((v) => !shouldRemoveKeys.includes(`${v}`));
    setSelectedSharedLinkKeys(newSelectedKeys);
  };

  const rowSelection: TableRowSelection<SharedLinkInfo> = {
    selectedRowKeys: selectedSharedLinkKeys,
    onChange: onSelectChange,
  };

  const [userInfo, visibleTableColumns, isMobile] = useStore((state) => [
    state.userInfo,
    state.visibleTableColumns,
    state.isMobile,
  ]);
  const [columns, setColumns] = useState<ColumnsType<SharedLinkInfo>>([]);
  useEffect(() => {
    const newColumns = getColumns({
      formatType: formatType,
      channelID: userInfo.channel_id,
      refresh: () => {
        console.log("refresh data!");
        refresh();
      },
    });

    setColumns(
      newColumns.filter((v) =>
        visibleTableColumns.includes(v.key as SharedLinkTableKey),
      ),
    );
  }, [userInfo.channel_id, formatType, visibleTableColumns]);

  const handleRefreshByFooter = () => {
    setSelectedSharedLinkKeys([]);
    refresh();
  };

  return (
    <>
      <Table
        rowSelection={rowSelection}
        columns={columns}
        dataSource={tableData}
        style={{
          marginTop: token.margin,
          marginBottom: isMobile ? "145px" : "65px",
        }}
        scroll={{
          x: isMobile ? "100%" : "0",
        }}
        loading={isLoading}
        pagination={{
          // showSizeChanger: true,
          total: totalSharedNum,
          defaultPageSize: 10,
          onChange: handlePageChange,
        }}
      />
      <SharedLinkTableFooter
        refresh={handleRefreshByFooter}
        selectedRows={selectedSharedLinks}
      />
    </>
  );
};

export default SharedLinkTable;
