import styled from "styled-components";
import { Link } from "react-router-dom";
import { Layout } from "antd";

const { Header } = Layout;

export const NavWrapper = styled(Header)`
  display: flex;
  justify-content: flex-end;
  align-items: center;
  padding: 25px 0;
  font-size: 16px;
  font-weight: 400;
  background-color: transparent;
`;

export const StyledLink = styled(Link)<{ color: React.CSSProperties["color"] }>`
  font-size: 16px;
  margin: 0 24px;
  color: ${({ color }) => color || "#000"};
`;
