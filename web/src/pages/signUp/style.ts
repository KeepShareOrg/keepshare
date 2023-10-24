import { Button, Form, Input } from "antd";
import styled from "styled-components";

export const Wrapper = styled.div`
  width: fit-content;
`;

export const StyledForm = styled(Form)`
  width: 300px;
  margin: 32px auto;
`;

export const StyledButton = styled(Button)`
  width: 100%;
  min-height: 40px;
`;

export const SignUpTips = styled.p`
  text-align: center;
  font-size: 14px;
`;

export const StyledInput = styled(Input)`
  height: 40px;
`;

export const PasswordInput = styled(Input.Password)`
  height: 40px;
`;

export const StyledItem = styled(Form.Item)`
  position: relative;
`;
