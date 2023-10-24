import { Button, Divider, Table, Typography, message, theme } from "antd";
import SubmissionResultHeader from "./SubmissionResultHeader";
import { useEffect, useState } from "react";
import { LinkFormatType, SharedLinkTableKey } from "@/constant";
import type { ColumnsType } from "antd/es/table";
import {
  getLinkToShareSubmissionResult,
  type SharedLinkInfo,
} from "@/api/link";
import { copyToClipboard, exportExcel, formatLinkWithType } from "@/util";
import useStore from "@/store";
import CopyOrOpenDropMenu from "../common/CopyOrOpenDropMenu";
import dayjs from "dayjs";
import { useNavigate } from "react-router-dom";
import { RoutePaths } from "@/router";

const { Paragraph } = Typography;

interface GetColumnsParams {
  formatType: LinkFormatType;
  channelID?: string;
}
const getColumns = ({
  formatType,
  channelID,
}: GetColumnsParams): ColumnsType<SharedLinkInfo> => [
  {
    title: SharedLinkTableKey.ORIGINAL_LINKS,
    key: SharedLinkTableKey.ORIGINAL_LINKS,
    ellipsis: true,
    render: ({ original_link }: SharedLinkInfo) => {
      const formatLink = formatLinkWithType(original_link, formatType);

      return <CopyOrOpenDropMenu state={"OK"} copy formatLink={formatLink} />;
    },
  },
  {
    title: SharedLinkTableKey.KEEP_SHARING_LINK,
    key: SharedLinkTableKey.KEEP_SHARING_LINK,
    ellipsis: true,
    render: ({ state, original_link }: SharedLinkInfo) => {
      const link = `${location.origin}/${channelID}/${encodeURIComponent(
        original_link,
      )}`;
      const formatLink = formatLinkWithType(link, formatType);

      return (
        <CopyOrOpenDropMenu state={state} open copy formatLink={formatLink} />
      );
    },
  },
  {
    title: SharedLinkTableKey.HOST_SHARED_LINK,
    key: SharedLinkTableKey.HOST_SHARED_LINK,
    ellipsis: true,
    render: ({ state, host_shared_link }: SharedLinkInfo) => {
      const formatLink = formatLinkWithType(host_shared_link, formatType);

      return (
        <CopyOrOpenDropMenu state={state} open copy formatLink={formatLink} />
      );
    },
  },
];

interface ComponentProps {
  links: string[];
}
const SubmissionResultTable = ({ links }: ComponentProps) => {
  const { token } = theme.useToken();

  const [tableData, setTableData] = useState<SharedLinkInfo[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  let totalSharedNum = 0;
  useEffect(() => {
    setIsLoading(true);
    getLinkToShareSubmissionResult(links)
      .then((data) => {
        const resultList = data.data?.list;
        if (resultList) {
          setTableData(resultList.map((v) => ({ ...v, key: `${v.auto_id}` })));
          totalSharedNum = resultList.length;
        }
      })
      .catch(({ error }) => {
        error && message.error("fetch submission result error!");
      })
      .finally(setIsLoading.bind(null, false));
  }, [links]);

  const userInfo = useStore((state) => state.userInfo);
  const [formatType, setFormatType] = useState<LinkFormatType>(
    LinkFormatType.TEXT,
  );
  const [columns, setColumns] = useState<ColumnsType<SharedLinkInfo>>([]);

  useEffect(
    () =>
      setColumns(
        getColumns({
          formatType,
          channelID: userInfo.channel_id,
        }),
      ),
    [userInfo, formatType],
  );

  const getExportData = (channelID: string) => {
    const exportData: unknown[][] = [
      [
        SharedLinkTableKey.ORIGINAL_LINKS,
        SharedLinkTableKey.KEEP_SHARING_LINK,
        SharedLinkTableKey.HOST_SHARED_LINK,
      ],
    ];

    tableData.forEach((v) => {
      exportData.push([
        v.original_link || "-",
        `${location.origin}/${channelID}/${encodeURIComponent(
          v.original_link,
        )}`,
        v.host_shared_link || "-",
      ]);
    });

    return exportData;
  };

  const handleCopy = () => {
    const exportData = getExportData(userInfo.channel_id || "-");

    const copyText = exportData.map((v) => v.join("\t")).join("\n");
    copyToClipboard(copyText);
  };

  const handleDownload = () => {
    const exportData = getExportData(userInfo.channel_id || "-");
    exportExcel(
      exportData,
      `submission-result-${dayjs().format("YYYY-MM-DD HH:mm:ss")}.xlsx`,
    );
  };

  /* view in shared links page, use filter by created_at between  */
  const getViewInSharedLinksSearchString = (data: SharedLinkInfo[]) => {
    const sortTimestamp = data.map((v) => dayjs(v.created_at).valueOf()).sort();
    const start = dayjs(sortTimestamp[0]).format("YYYY-MM-DD HH:mm:ss");
    const end = dayjs(sortTimestamp[sortTimestamp.length - 1]).format(
      "YYYY-MM-DD HH:mm:ss",
    );

    let result = "";
    if (sortTimestamp.length > 0) {
      sortTimestamp.length === 1
        ? (result = `created_at="${start}"`)
        : (result = `created_at>="${start}" created_at<="${end}"`);
    }
    return window.encodeURIComponent(result);
  };

  const navigate = useNavigate();
  const handleViewInSharedLinks = () => {
    const search = getViewInSharedLinksSearchString(tableData);
    navigate(
      search
        ? `${RoutePaths.SharedLinks}?search=${search}`
        : RoutePaths.SharedLinks,
    );
  };

  return (
    <Paragraph style={{ marginTop: token.marginXL }}>
      <SubmissionResultHeader
        handleCopy={handleCopy}
        handleDownload={handleDownload}
        handleFormatChange={setFormatType}
      />
      <Table
        style={{ marginTop: token.margin }}
        columns={columns}
        dataSource={tableData}
        loading={isLoading}
        pagination={{
          total: totalSharedNum,
          defaultPageSize: 10,
          pageSizeOptions: ["2", "10", "50", "100"],
        }}
      />
      <Divider />
      <Paragraph>
        <Button
          type="link"
          style={{ color: token.colorPrimaryHover }}
          onClick={handleViewInSharedLinks}
        >
          View in Shared Links
        </Button>
      </Paragraph>
    </Paragraph>
  );
};

export default SubmissionResultTable;
