import { Form, Input, Button } from "antd";
import styled from "styled-components";

export const StyledForm = styled(Form)`
  max-width: 360px;
  margin: 32px auto;
`;

export const StyledInput = styled(Input)`
  height: 40px;
`;

export const PasswordInput = styled(Input.Password)`
  height: 40px;
`;

export const TextLink = styled(Button)<{ color?: string; padding?: number; }>`
  color: ${props => props.color};
  padding: 0 ${props => props.padding}px;
  height: auto;
`
