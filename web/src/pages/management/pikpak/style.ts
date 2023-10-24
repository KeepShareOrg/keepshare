import { Statistic } from "antd";
import styled from "styled-components";

export const RewardStatistic = styled(Statistic)<{
  color: React.CSSProperties["color"];
}>`
  .ant-statistic-title {
    color: ${({ color }) => color || "#000"};
  }
`;
