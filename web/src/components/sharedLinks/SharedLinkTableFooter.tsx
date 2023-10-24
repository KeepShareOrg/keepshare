import {
  addToBlacklist,
  removeFromBlacklist,
  type SharedLinkInfo,
} from "@/api/link";
import useStore from "@/store";
import {
  Badge,
  Button,
  Checkbox,
  Col,
  message,
  Modal,
  Row,
  Space,
  theme,
  Typography,
} from "antd";
import React, { useEffect, useRef, useState } from "react";
import { TableFooterWrapper } from "./style";
import {
  DownloadOutlined,
  MinusCircleOutlined,
  PlusCircleOutlined,
} from "@ant-design/icons";
import { SharedLinkTableKey, supportTableColumns } from "@/constant";
import ExportColumnDialog, {
  ExportColumnDialogRef,
} from "./ExportColumnDialog";
import { copyToClipboard, exportExcel, formatBytes } from "@/util";
import { match } from "ts-pattern";
import dayjs from "dayjs";

const { Text } = Typography;

// transform table data to export excel
const transformTableData = (
  data: SharedLinkInfo,
  columns: SharedLinkTableKey[],
  channelID: string,
) => {
  const result: Partial<Record<SharedLinkTableKey, unknown>> = {};

  columns.forEach((column) => {
    result[column] = match(column)
      .with(SharedLinkTableKey.KEEP_SHARING_LINK, () => {
        return `${location.origin}/${channelID}/${encodeURIComponent(
          data.original_link,
        )}`;
      })
      .with(SharedLinkTableKey.TITLE, () => data.title)
      .with(SharedLinkTableKey.HOST_SHARED_LINK, () => data.host_shared_link)
      .with(SharedLinkTableKey.CREATED_AT, () =>
        dayjs(data.created_at).format("YYYY-MM-DD HH:mm"),
      )
      .with(SharedLinkTableKey.VISITOR, () => data.visitor)
      .with(SharedLinkTableKey.STORED, () => data.stored)
      .with(SharedLinkTableKey.SIZE, () =>
        typeof data.size === "string" ? data.size : formatBytes(data.size),
      )
      .with(SharedLinkTableKey.DAYS_NOT_VISIT, () => data.days_not_visit || "-")
      .with(SharedLinkTableKey.STATE, () => data.state)
      .with(SharedLinkTableKey.ORIGINAL_LINKS, () => data.original_link)
      .with(SharedLinkTableKey.ACTION, () => "")
      .with(SharedLinkTableKey.CREATED_BY, () => data.created_by)
      .exhaustive();
  });

  return result;
};

interface ComponentProps {
  refresh: VoidFunction;
  selectedRows: SharedLinkInfo[];
}
const SharedLinkTableFooter = ({ refresh, selectedRows }: ComponentProps) => {
  const { token } = theme.useToken();
  const [asideCollapsed, userInfo, isMobile, totalSharedLinks] = useStore(
    (state) => [
      state.asideCollapsed,
      state.userInfo,
      state.isMobile,
      state.totalSharedLinks,
    ],
  );

  const [buttonState, setButtonState] = useState<
    "show-add-blacklist" | "show-remove-blacklist" | "just-copy-download"
  >("just-copy-download");

  useEffect(() => {
    const haveValidRow = selectedRows.some((v) => v.state === "OK");
    const haveBlackListRow = selectedRows.some((v) => v.state === "BLOCKED");

    if (haveValidRow && haveBlackListRow) {
      setButtonState("just-copy-download");
    } else if (haveValidRow) {
      setButtonState("show-add-blacklist");
    } else if (haveBlackListRow) {
      setButtonState("show-remove-blacklist");
    }
  }, [selectedRows]);

  const [showAddConfirm, setShowAddConfirm] = useState(false);
  const handleAddToBlacklist = async () => {
    try {
      const links = selectedRows.map((v) => v.original_link);
      setShowAddConfirm(false);
      const { error } = await addToBlacklist(links);
      if (error) {
        message.error(error.message);
        return;
      }
      refresh();
      message.success("add to blacklist success!");
    } catch (err) {
      console.error("add to blacklist error: ", err);
    }
  };

  // batch remove black list
  const [showRemoveConfirm, setShowRemoveConfirm] = useState(false);
  const [isDeleteHostLinkAndFile, setIsDeleteHostLinkAndFile] = useState(false);
  const handleRemoveFromBlacklist = async () => {
    try {
      const links = selectedRows.map((v) => v.original_link);
      setShowRemoveConfirm(false);
      const { error } = await removeFromBlacklist(
        links,
        isDeleteHostLinkAndFile,
      );
      if (error) {
        message.error(error.message);
        return;
      }
      refresh();
      message.success("remove from blacklist success!");
    } catch (err) {
      console.error("remove from blacklist error: ", err);
    }
  };

  const getExportData = (
    selectedRows: SharedLinkInfo[],
    selectedColumns: SharedLinkTableKey[],
    channelID: string,
    includeHeader: boolean,
  ) => {
    const exportData: unknown[][] = [];
    if (includeHeader) {
      exportData.push(Object.keys(selectedRows[0]));
    }
    selectedRows.forEach((rowData) => {
      const rowRet = transformTableData(rowData, selectedColumns, channelID);
      exportData.push(Object.values(rowRet));
    });
    return exportData;
  };

  const exportRef = useRef<ExportColumnDialogRef>(null);
  const handleCopyAction = (
    selectedColumns: React.Key[],
    includeHeader: boolean,
  ) => {
    exportRef.current?.hideModel();
    const exportData = getExportData(
      selectedRows,
      selectedColumns as SharedLinkTableKey[],
      userInfo.channel_id || "-",
      includeHeader,
    );

    const copyText = exportData.map((v) => v.join("\t")).join("\n");
    copyToClipboard(copyText);
    message.success("copy success!");
  };
  const handleCopy = () => {
    exportRef.current?.showModel({
      title: "Select columns to copy",
      footerButtonRender: function (
        selectedColumns: React.Key[],
        includeHeader: boolean,
      ): React.ReactNode {
        return (
          <Space>
            <Button onClick={() => exportRef.current?.hideModel()}>
              cancel
            </Button>
            <Button
              type="primary"
              onClick={() => handleCopyAction(selectedColumns, includeHeader)}
            >
              copy
            </Button>
          </Space>
        );
      },
    });
  };

  const handleDownloadAction = (
    selectedColumns: React.Key[],
    includeHeader: boolean,
  ) => {
    exportRef.current?.hideModel();
    const exportData = getExportData(
      selectedRows,
      selectedColumns as SharedLinkTableKey[],
      userInfo.channel_id || "-",
      includeHeader,
    );

    exportExcel(
      exportData,
      `shared-links-${dayjs().format("YYYY-MM-DD HH:mm:ss")}.xlsx`,
    );
    message.success("export success!");
  };
  const handleDownload = () => {
    exportRef.current?.showModel({
      title: "Select columns in CSV",
      footerButtonRender: function (
        selectedColumns: React.Key[],
        includeHeader: boolean,
      ): React.ReactNode {
        return (
          <Space>
            <Button onClick={() => exportRef.current?.hideModel()}>
              Cancel
            </Button>
            <Button
              type="primary"
              icon={<DownloadOutlined />}
              onClick={() =>
                handleDownloadAction(selectedColumns, includeHeader)
              }
            >
              Download CSV
            </Button>
          </Space>
        );
      },
    });
  };

  const [dataCrossPage, setDataCrossPage] = useState(0);
  useEffect(() => {
    const pageSize = 10;
    const totalPage = Math.ceil(totalSharedLinks.length / pageSize);
    const selectedKeys = selectedRows.map((v) => `${v.auto_id}`);
    let totalCrossPage = 0;
    for (let i = 0; i < totalPage; i++) {
      const pageData = totalSharedLinks.slice(i * pageSize, (i + 1) * pageSize);
      for (let v = 0; v < pageData.length; v++) {
        if (selectedKeys.includes(`${pageData[v].auto_id}`)) {
          totalCrossPage++;
          break;
        }
      }
    }
    setDataCrossPage(totalCrossPage);
  }, [selectedRows]);

  return (
    <>
      <ExportColumnDialog columns={supportTableColumns} ref={exportRef} />
      {selectedRows.length > 0 && (
        <TableFooterWrapper
          style={{
            width: isMobile
              ? "100%"
              : asideCollapsed
              ? "calc(100% - 80px)"
              : "calc(100% - 200px)",
            background: token.colorBgBase,
            borderTop: `1px solid ${token.colorBorder}`,
            paddingBlock: isMobile ? token.padding : 0,
            paddingInline: isMobile ? token.padding : token.paddingXL,
          }}
        >
          <Row align="middle" style={{ width: "100%" }}>
            <Col xs={24} md={8}>
              <Badge color={token.colorPrimary} count={selectedRows.length} />
              <Text style={{ marginLeft: token.marginXS }}>
                {`Selected in ${dataCrossPage} pages`}
              </Text>
            </Col>
            <Col xs={24} md={16}>
              <Space
                wrap
                style={{
                  width: "100%",
                  marginTop: isMobile ? token.margin : 0,
                }}
                align="center"
              >
                {buttonState === "show-remove-blacklist" && (
                  <Button
                    icon={<MinusCircleOutlined />}
                    onClick={() => setShowRemoveConfirm(true)}
                  >
                    Remove from Blacklist
                  </Button>
                )}
                {buttonState === "show-add-blacklist" && (
                  <Button
                    icon={<PlusCircleOutlined />}
                    onClick={() => setShowAddConfirm(true)}
                  >
                    Add to Blacklist
                  </Button>
                )}
                <Button onClick={handleCopy}>Copy</Button>
                <Button
                  type="primary"
                  icon={<DownloadOutlined />}
                  onClick={handleDownload}
                >
                  Download CSV
                </Button>
              </Space>
            </Col>
          </Row>
        </TableFooterWrapper>
      )}

      <Modal
        title="Remove from Blacklist"
        open={showRemoveConfirm}
        onOk={handleRemoveFromBlacklist}
        onCancel={() => setShowRemoveConfirm(false)}
      >
        <Text>
          After remove, the current link will automatically jump to the shared
          link of the host drive.
        </Text>
      </Modal>

      <Modal
        title="Add link to blacklist"
        open={showAddConfirm}
        onOk={handleAddToBlacklist}
        onCancel={() => {
          setShowAddConfirm(false);
          setIsDeleteHostLinkAndFile(false);
        }}
      >
        <Space>
          <Text>
            After adding to the blacklist, the Keep Sharing Link will not be
            able to jump to Host Shared Link
          </Text>
        </Space>
        <Space style={{ marginTop: token.marginSM }}>
          <Checkbox
            checked={isDeleteHostLinkAndFile}
            onChange={() =>
              setIsDeleteHostLinkAndFile(!isDeleteHostLinkAndFile)
            }
            style={{ color: token.colorTextSecondary }}
          >
            Delete the Host Shared Link and files
          </Checkbox>
        </Space>
      </Modal>
    </>
  );
};

export default SharedLinkTableFooter;
