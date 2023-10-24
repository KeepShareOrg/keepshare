import { DownOutlined } from "@ant-design/icons";
import {
  Typography,
  Col,
  Divider,
  Dropdown,
  InputNumber,
  Row,
  Select,
  type SelectProps,
  Space,
  type MenuProps,
  theme,
} from "antd";
import {
  type HandleClickType,
  SupportOperatorLabels,
  SupportOperatorValueMap,
  transferSelectionToQueryCondition,
  ValueSupportOperatorMap,
} from "./filter.script";
import type { QueryCondition, BasicItemType } from "./FilterItem";
import { useEffect, useState } from "react";
import { type DefaultOptionType } from "antd/es/select";
import useStore from "@/store";

const { Text } = Typography;

export interface MultipleFilterItemType extends BasicItemType {
  type: "multiple";
  operators: SelectProps["options"];
  defaultOperator: SupportOperatorLabels;
  getSelections: (handleClick: HandleClickType) => MenuProps["items"];
}
export const MultipleFilterItem = ({
  title,
  unit,
  searchKey,
  filters,
  operators,
  defaultOperator,
  getSelections,
  handleFilterChange,
}: MultipleFilterItemType) => {
  const [selectedOperator, setSelectedOperator] =
    useState<SupportOperatorLabels>(defaultOperator);

  const handleSelect = (
    _: SupportOperatorLabels,
    item: DefaultOptionType | DefaultOptionType[],
  ) => {
    Array.isArray(item) ||
      setSelectedOperator(item.label as SupportOperatorLabels);
  };

  const [suffixUnit, setSuffixUnit] = useState<string>("");
  const [startValue, setStartValue] = useState<number>();
  const [endValue, setEndValue] = useState<number>();
  const [singleValue, setSingleValue] = useState<number>();

  useEffect(() => {
    const conditions = filters?.filter(({ key }) => key === searchKey);
    if (!conditions || conditions.length === 0) {
      setStartValue(undefined);
      setEndValue(undefined);
      setSingleValue(undefined);
      setSelectedOperator(ValueSupportOperatorMap["*"]);
      return;
    }
    if (conditions.length > 1) {
      const lessCondition = conditions.find(({ operator }) => operator === "<");
      const grateCondition = conditions.find(
        ({ operator }) => operator === ">",
      );

      const unit = lessCondition?.unit || grateCondition?.unit;
      unit && setSuffixUnit(unit);

      if (lessCondition && grateCondition) {
        setStartValue(Number(grateCondition?.value));
        setEndValue(Number(lessCondition?.value));
        setSingleValue(undefined);
        setSelectedOperator(ValueSupportOperatorMap["between"]);
      }
    } else {
      setSingleValue(Number(conditions[0].value));
      setStartValue(undefined);
      setEndValue(undefined);
      setSelectedOperator(
        ValueSupportOperatorMap[conditions[0].operator] ||
          SupportOperatorLabels.ANY,
      );
    }
  }, [filters]);

  useEffect(() => {
    const operator = SupportOperatorValueMap[selectedOperator];
    const result: QueryCondition = { key: searchKey, operator, value: 0 };
    if (unit || suffixUnit) {
      result.unit = unit || suffixUnit;
    }
    if (selectedOperator === SupportOperatorLabels.ANY) {
      result.value = "";
    }

    const shouldUpdateFilters =
      singleValue !== undefined ||
      (startValue !== undefined && endValue !== undefined);
    if (!shouldUpdateFilters) {
      return;
    }
    if (selectedOperator === SupportOperatorLabels.BETWEEN) {
      result.value = [startValue!, endValue!];
    } else {
      result.value = singleValue!;
    }
    handleFilterChange?.(result);
  }, [selectedOperator, singleValue, startValue, endValue]);

  const handleSelectionClick = (v: string) => {
    const condition = transferSelectionToQueryCondition(v);
    const operatorLabel = ValueSupportOperatorMap[condition.operator];
    operatorLabel && setSelectedOperator(operatorLabel);

    if (
      operatorLabel === SupportOperatorLabels.BETWEEN &&
      Array.isArray(condition.value)
    ) {
      setStartValue(Number(condition.value?.[0]));
      setEndValue(Number(condition.value?.[1]));
    } else {
      setSingleValue(Number(condition.value));
    }
  };

  const isMobile = useStore((state) => state.isMobile);
  const { token } = theme.useToken();

  return (
    <Row align="middle">
      <Col xs={24} md={5}>
        <Text strong>{title}</Text>
      </Col>
      <Col xs={24} md={19} style={{ marginTop: token.marginXS }}>
        <Space split wrap>
          <Select
            value={selectedOperator}
            defaultValue={selectedOperator}
            bordered={false}
            popupMatchSelectWidth={false}
            options={operators}
            onChange={handleSelect}
          />
          {selectedOperator !== SupportOperatorLabels.ANY &&
            (selectedOperator === SupportOperatorLabels.BETWEEN ? (
              <Space>
                <InputNumber
                  style={{ width: isMobile ? "72px" : "100px" }}
                  value={startValue}
                  suffix={unit || suffixUnit}
                  onChange={(v) => setStartValue(v || 0)}
                  placeholder="Min"
                ></InputNumber>
                -
                <InputNumber
                  style={{ width: isMobile ? "72px" : "100px" }}
                  value={endValue}
                  suffix={unit || suffixUnit}
                  onChange={(v) => setEndValue(v || 0)}
                  placeholder="Max"
                  min={startValue}
                ></InputNumber>
              </Space>
            ) : (
              <InputNumber
                style={{ width: isMobile ? "72px" : "100px" }}
                value={singleValue}
                suffix={unit || suffixUnit}
                onChange={(v) => setSingleValue(v || 0)}
                placeholder="Enter item"
              ></InputNumber>
            ))}
          <Divider type="vertical" />
          <Dropdown
            menu={{ items: getSelections(handleSelectionClick) }}
            trigger={["click"]}
          >
            <Space style={{ cursor: "pointer" }}>
              <Text>By Selection</Text>
              <DownOutlined style={{ color: token.colorTextTertiary }} />
            </Space>
          </Dropdown>
        </Space>
      </Col>
    </Row>
  );
};
