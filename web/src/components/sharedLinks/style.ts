import styled from "styled-components";
import { Tag } from "antd";

const { CheckableTag } = Tag;
export const SelectionWrapper = styled.div`
  display: flex;
  align-items: center;
`;

export const IconImg = styled.img`
  display: inline-block;
  width: 18px;
  height: 18px;
  margin-left: 24px;
  margin-right: 6px;
  object-fit: fill;
`;

export const MultipleWrapper = styled.div`
  margin-top: 16px;
`;

export const AdvancedPaneWrapper = styled.div`
  width: 747px;
`;

export const TableCellIcon = styled.img`
  width: 16px;
  height: 16px;
  margin-right: 4px;
  align-self: center;
`;

export const TableFooterWrapper = styled.div`
  display: flex;
  align-items: center;
  position: fixed;
  width: calc(100% - 200px);
  min-height: 64px;
  right: 0;
  bottom: 0;
  transition: width 0.2s ease;
`;

export const CustomCheckableTag = styled(CheckableTag)`
  &.ant-tag-checkable-checked {
    background-color: #f0f0f0;
    color: #000;
  }
`;
