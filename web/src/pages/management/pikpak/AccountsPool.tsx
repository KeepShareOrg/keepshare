import {
  clearPikPakAccountStorage,
  ClearPikPakAccountStorageParams,
  getPikPakAccountStatistics,
} from "@/api/pikpak";
import { RoutePaths } from "@/router";
import useStore from "@/store";
import { formatBytes } from "@/util";
import { DeleteOutlined, InfoCircleOutlined } from "@ant-design/icons";
import {
  Button,
  Checkbox,
  Divider,
  InputNumber,
  Space,
  Switch,
  Typography,
  message,
  theme,
} from "antd";
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
const { Paragraph, Title, Text } = Typography;

const getBasicUsageInfo = async (): Promise<React.ReactNode[]> => {
  try {
    const { data, error } = await getPikPakAccountStatistics({
      host: "pikpak",
      stored_count_lt: [1, 10, 100, 1000],
      not_stored_days_gt: [60, 180, 365],
    });

    if (error) {
      message.error(error.message);
      return [];
    }

    return [
      <Space direction="vertical">
        {data?.stored_count_lt.map(({ number, total_count, total_size }, i) => {
          const content =
            number === 1
              ? `Links that have never been stored, total  ${total_count} links · ${formatBytes(
                  total_size,
                )}`
              : `Links stored less than ${number} times,  total ${total_count} links · ${formatBytes(
                  total_size,
                )}`;
          return <Text key={`stored_count_lt_${i}`}>{content}</Text>;
        })}
      </Space>,
      <Space direction="vertical" style={{ marginTop: "12px" }}>
        {data?.not_stored_days_gt.map(
          ({ number, total_count, total_size }, i) => {
            const content = `Links are not stored for greater than ${number} days , total ${total_count} links · ${formatBytes(
              total_size,
            )}`;
            return <Text key={`not_stored_days_gt_${i}`}>{content}</Text>;
          },
        )}
      </Space>,
    ];
  } catch (err) {
    console.error("getBasicUsageInfo error: ", err);
    return [];
  }
};

const AccountPool = () => {
  const { token } = theme.useToken();

  const [info, isMobile] = useStore((state) => [
    state.pikPakInfo,
    state.isMobile,
  ]);

  const [basicUsageInfo, setBasicUsageInfo] = useState<React.ReactNode[]>([]);
  // get basic usage info data for all links block
  useEffect(() => {
    getBasicUsageInfo().then(setBasicUsageInfo);
  }, []);

  const [clearParams, setClearParams] =
    useState<ClearPikPakAccountStorageParams>({
      host: "pikpak",
      stored_count_lt: 100,
      not_stored_days_gt: 60,
      only_for_premium: false,
    });

  type ConditionValueType = Record<
    "neverStored" | "storedLessTimes" | "storedGraterDays" | "onlyClearPremium",
    boolean
  >;
  const [deleteCondition, setDeleteCondition] = useState<ConditionValueType>({
    neverStored: true,
    storedLessTimes: true,
    storedGraterDays: false,
    onlyClearPremium: true,
  });

  const [isDeleting, setIsDeleting] = useState(false);
  const handleDeleteLinks = async () => {
    try {
      setIsDeleting(true);
      const params: ClearPikPakAccountStorageParams = {
        host: "pikpak",
        only_for_premium: deleteCondition.onlyClearPremium,
      };

      if (deleteCondition.neverStored) {
        params.stored_count_lt = 1;
      }

      if (deleteCondition.storedLessTimes) {
        params.stored_count_lt = clearParams.stored_count_lt;
      }

      if (deleteCondition.storedGraterDays) {
        params.not_stored_days_gt = clearParams.not_stored_days_gt;
      }
      const { error } = await clearPikPakAccountStorage(params);
      if (error) {
        message.error(error.message);
        return;
      }
      message.success("success!");
    } catch (err) {
      console.error("handle delete links error: ", err);
    } finally {
      setIsDeleting(false);
    }
  };

  // real-time statistics data
  const [realTimeStatistics, setRealTimeStatistics] = useState({
    lessTimes: {
      number: 100,
      total_count: 0,
      total_size: 0,
    },
    graterDays: {
      number: 60,
      total_count: 0,
      total_size: 0,
    },
  });

  useEffect(() => {
    getPikPakAccountStatistics({
      host: "pikpak",
      stored_count_lt: [clearParams.stored_count_lt || 100],
      not_stored_days_gt: [clearParams.not_stored_days_gt || 60],
    }).then(({ data, error }) => {
      if (error === null && data) {
        setRealTimeStatistics({
          lessTimes: data.stored_count_lt[0],
          graterDays: data.not_stored_days_gt[0],
        });
      }
    });
  }, [clearParams.stored_count_lt, clearParams.not_stored_days_gt]);

  return (
    <>
      <Space align="start" wrap>
        <Space style={{ width: "200px", marginRight: token.marginLG }}>
          <Text style={{ color: token.colorTextSecondary }}>Free Accounts</Text>
        </Space>
        <Space direction="vertical">
          <Text>
            The number of Free Accounts used is
            <Text strong>{info.workers?.free.count}</Text>, and the number of
            accounts is <Text strong>unlimited.</Text>
          </Text>
          <Text>
            The storage used is
            <Text strong>{formatBytes(info.workers?.free.used || 0)}</Text>, and
            the storage is <Text strong>unlimited.</Text>
          </Text>
        </Space>
      </Space>
      <Divider />
      <Space align="start" wrap>
        <Space style={{ width: "200px", marginRight: token.marginLG }}>
          <Text style={{ color: token.colorTextSecondary }}>
            Donated Premium Accounts
          </Text>
        </Space>
        <Space direction="vertical">
          <Text>
            The number of Free Accounts used is
            <Text strong>{info.workers?.premium.count}</Text>, and the number of
            accounts is <Text strong>unlimited.</Text>
          </Text>
          <Text>
            The storage used is
            <Text strong>{formatBytes(info.workers?.premium.used || 0)}</Text>,
            and the storage is <Text strong>unlimited.</Text>
          </Text>
          <Space style={{ marginTop: token.margin }}>
            <InfoCircleOutlined />
            <Text style={{ color: token.colorTextSecondary }}>
              To avoid account abuse, KeepShare provides each webmaster with up
              to 10 Premium sub-accounts, which are automatically assigned when
              needed
            </Text>
          </Space>
        </Space>
      </Space>
      <Divider />
      <Space align="start" wrap>
        <Space
          style={{
            width: isMobile ? "auto" : "200px",
            marginRight: token.marginLG,
          }}
        >
          <Text style={{ color: token.colorTextSecondary }}>
            Release Premium Accounts Storage
          </Text>
        </Space>
        <Space
          direction="vertical"
          style={{ width: isMobile ? "auto" : "600px" }}
        >
          <Text>
            The PikPak Premium account assigned by KeepShare to each account is
            limited. When sharing files fails due to insufficient storage, you
            can delete some share links and try again.
          </Text>
          <Title level={5} style={{ marginTop: token.marginLG }}>
            All Links
          </Title>

          {...basicUsageInfo}

          <Title level={5} style={{ marginTop: token.marginLG }}>
            Delete Links
          </Title>
          <Checkbox
            checked={deleteCondition.neverStored}
            onChange={() =>
              setDeleteCondition({
                ...deleteCondition,
                neverStored: !deleteCondition.neverStored,
              })
            }
          >
            Never stored
          </Checkbox>
          <Space wrap>
            <Checkbox
              checked={deleteCondition.storedLessTimes}
              onChange={() =>
                setDeleteCondition({
                  ...deleteCondition,
                  storedLessTimes: !deleteCondition.storedLessTimes,
                })
              }
            />
            <Text>Stored less than</Text>
            <InputNumber
              value={clearParams.stored_count_lt}
              min={0}
              onChange={(v) =>
                setClearParams({ ...clearParams, stored_count_lt: Number(v) })
              }
            ></InputNumber>
            <Text>{`times Stored less than, total ${
              realTimeStatistics.lessTimes.total_count
            } links · ${formatBytes(
              realTimeStatistics.lessTimes.total_size,
            )}`}</Text>
          </Space>
          <Space wrap>
            <Checkbox
              checked={deleteCondition.storedGraterDays}
              onChange={() =>
                setDeleteCondition({
                  ...deleteCondition,
                  storedGraterDays: !deleteCondition.storedGraterDays,
                })
              }
            />
            <Text>Stored for greater than</Text>
            <InputNumber
              min={0}
              value={clearParams.not_stored_days_gt}
              onChange={(v) =>
                setClearParams({
                  ...clearParams,
                  not_stored_days_gt: Number(v),
                })
              }
              style={{ minWidth: "60px" }}
            ></InputNumber>
            <Text>{`days ago, total ${
              realTimeStatistics.graterDays.total_count
            } links · ${formatBytes(
              realTimeStatistics.graterDays.total_size,
            )}`}</Text>
          </Space>
          <Space>
            <Text> Only delete the links of Donated Premium Accounts </Text>
            <Switch
              checked={deleteCondition.onlyClearPremium}
              onChange={() =>
                setDeleteCondition({
                  ...deleteCondition,
                  onlyClearPremium: !deleteCondition.onlyClearPremium,
                })
              }
            />
          </Space>
          <Divider />
          <Space
            align="baseline"
            style={{ width: "100%", justifyContent: "space-between" }}
          >
            <Paragraph>
              <Link
                style={{ color: token.colorPrimaryHover }}
                to={RoutePaths.SharedLinks}
              >
                View in Shared Links
              </Link>
            </Paragraph>
            <Space>
              <Button
                type="primary"
                loading={isDeleting}
                onClick={handleDeleteLinks}
                icon={<DeleteOutlined />}
              >
                Delete
              </Button>
            </Space>
          </Space>
        </Space>
      </Space>
    </>
  );
};

export default AccountPool;
