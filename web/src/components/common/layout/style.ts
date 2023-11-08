import styled from "styled-components";
import { Layout } from "antd";

const { Header, Content } = Layout;

export const NormalLayout = styled(Layout)`
  width: 100%;
  height: 100%;
  background-color: #fff;
`

export const NormalHeader = styled(Header)`
  background-color: #fff;
  padding: 12px 24px;
  height: auto;
  display: flex;
`

export const NormalContent = styled(Content)`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
`

export const LogoBox = styled.div`
  display: flex;
`

export const LogoImg = styled.img`
  width: 208px;
`
