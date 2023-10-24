import { type SharedLinkInfo } from "@/api/link";
import { Dropdown, Space, Typography, message } from "antd";
import { CopyOutlined, LinkOutlined, LoadingOutlined } from "@ant-design/icons";
import { copyToClipboard } from "@/util";

const { Paragraph, Text } = Typography;

export type LinkActionType = "Copy" | "Open";
export const getLinkActionMenuItems = (
  actions: LinkActionType[],
  link: string,
) => {
  return actions.map((action) => {
    return {
      key: action,
      icon: action === "Copy" ? <CopyOutlined /> : <LinkOutlined />,
      label: action,
      onClick: () => {
        if (action === "Copy") {
          copyToClipboard(link);
          message.success("copy success!");
        }
        if (action === "Open") {
          window.open(link);
        }
      },
    };
  });
};

interface ComponentProps {
  copy?: boolean;
  open?: boolean;
  state: SharedLinkInfo["state"];
  formatLink: string;
  children?: React.ReactNode;
}
const CopyOrOpenDropMenu = ({
  copy,
  open,
  state,
  formatLink,
  children,
}: ComponentProps) => {
  const processing = ["CREATED", "PENDING"].includes(state);

  const actionTypes: LinkActionType[] = [];
  if (state === "OK") {
    open && actionTypes.push("Open");
    copy && actionTypes.push("Copy");
  }

  return (
    <Dropdown
      key={formatLink}
      trigger={actionTypes.length > 0 ? ["hover"] : []}
      menu={{ items: getLinkActionMenuItems(actionTypes, formatLink) }}
      overlayStyle={{ minWidth: "145px" }}
    >
      {state === "OK" ? (
        <Paragraph style={{ display: "flex", width: "100%", marginBottom: 0 }}>
          {children}
          <Text ellipsis title={formatLink}>
            {formatLink}
          </Text>
        </Paragraph>
      ) : (
        <Space>
          {processing && <LoadingOutlined />}
          {processing ? (
            <Text ellipsis>PROCESSING</Text>
          ) : (
            <Text ellipsis type="danger">
              {state}
            </Text>
          )}
        </Space>
      )}
    </Dropdown>
  );
};

export default CopyOrOpenDropMenu;
