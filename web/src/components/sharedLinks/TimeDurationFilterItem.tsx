import { Col, Row, Typography, DatePicker, theme } from "antd";
import dayjs from "dayjs";
import type { BasicItemType } from "./FilterItem";
import {
  SupportOperatorLabels,
  SupportOperatorValueMap,
} from "./filter.script";
import { useEffect, useState } from "react";
import { RangePickerProps } from "antd/es/date-picker";
import useStore from "@/store";

const { Text } = Typography;
const { RangePicker } = DatePicker;

export interface TimeDurationFilterItemType extends BasicItemType {
  type: "time-duration";
}
export const TimeDurationFilterItem = ({
  title,
  searchKey,
  filters,
  handleFilterChange,
}: TimeDurationFilterItemType) => {
  const dataFormat = "YYYY-MM-DD HH:mm:ss";

  const isMobile = useStore((state) => state.isMobile);
  const { token } = theme.useToken();

  const [dataRange, setDataRange] = useState<RangePickerProps["value"]>();

  // eslint-disable-next-line
  const handleDateChange = (_: any, [s, e]: string[]) => {
    const operator = SupportOperatorValueMap[SupportOperatorLabels.BETWEEN];
    if (!s || !e) {
      setDataRange(undefined);
      handleFilterChange?.({ key: searchKey, operator: "*", value: [s, e] });
    } else {
      setDataRange([dayjs(s), dayjs(e)]);
      handleFilterChange?.({ key: searchKey, operator, value: [s, e] });
    }
  };

  useEffect(() => {
    const value = filters
      ?.filter(({ key }) => key === searchKey)
      .map((v) => v.value) as string[];
    if (
      Array.isArray(value) &&
      value.length === 2 &&
      value.every((v) => typeof v === "string")
    ) {
      setDataRange([dayjs(value[0]), dayjs(value[1])]);
    }
  }, [filters]);

  return (
    <Row align="middle">
      <Col xs={24} md={5}>
        <Text strong>{title}</Text>
      </Col>
      <Col xs={24} md={10} style={{ marginTop: isMobile ? token.marginXS : 0 }}>
        <RangePicker
          allowEmpty={[true, true]}
          placeholder={["Created start date", "Created end date"]}
          value={dataRange}
          showTime={{ format: "HH:mm:ss", defaultValue: [] }}
          format={dataFormat}
          presets={[
            {
              label: "Yesterday",
              value: [
                dayjs(dayjs().add(-1, "d").format("YYYY-MM-DD 00:00:00")),
                dayjs(dayjs().add(-1, "d").format("YYYY-MM-DD 23:59:59")),
              ],
            },
            { label: "Within a week", value: [dayjs().add(-7, "d"), dayjs()] },
            {
              label: "Within a month",
              value: [dayjs().add(-1, "month"), dayjs()],
            },
            { label: "Last 7 Days", value: [dayjs().add(-7, "d"), dayjs()] },
            { label: "Last 14 Days", value: [dayjs().add(-14, "d"), dayjs()] },
            { label: "Last 30 Days", value: [dayjs().add(-30, "d"), dayjs()] },
            { label: "Last 90 Days", value: [dayjs().add(-90, "d"), dayjs()] },
          ]}
          onChange={handleDateChange}
        />
      </Col>
    </Row>
  );
};
