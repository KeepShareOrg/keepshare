import { ResetPasswordSteps, StepComponentParams } from "@/constant";
import { StyledButton, StyledForm, StyledInput } from "@/pages/signUp/style";
import { type AlertProps, theme, Form, Alert } from "antd";
import { useState } from "react";
import { useTranslation } from "react-i18next";

interface FieldType {
  password?: string;
  confirmPassword?: string;
}
type ErrorMessage = string;
const validateFormFailed = ({
  password,
  confirmPassword,
}: FieldType): ErrorMessage => {
  console.log(password, confirmPassword);
  if (password?.trim() === "" || confirmPassword?.trim() === "") {
    return "password or confirm password is required";
  }

  if (password && confirmPassword && password !== confirmPassword) {
    return "password and confirm password do not match";
  }

  return "";
};

// Take the third step to reset the password and set a new password
const StepEnterNewPassword = ({ setStep }: StepComponentParams) => {
  const { t } = useTranslation();
  const [form] = Form.useForm<{
    password?: string;
    confirmPassword?: string;
  }>();

  const { token } = theme.useToken();

  const [formMessage, setFormMessage] = useState<{
    type: AlertProps["type"];
    message: string;
  }>({
    type: "error",
    message: "",
  });

  const handleResetPassword = () => {
    setStep(ResetPasswordSteps.RESET_PASSWORD_RESULT);
  };

  const handleInputPassword = () => {
    const validateResultMessage = validateFormFailed(form.getFieldsValue());
    setFormMessage({ type: "error", message: validateResultMessage });
  };

  return (
    <StyledForm
      form={form}
      layout="vertical"
      onFinish={handleResetPassword}
      validateTrigger={[]}
      autoComplete="off"
    >
      <Form.Item
        name="password"
        label="Please set a new password"
        style={{ marginBottom: token.marginSM }}
      >
        <StyledInput
          placeholder="Verification code"
          onChange={handleInputPassword}
        />
      </Form.Item>
      <Form.Item name="confirmPassword">
        <StyledInput placeholder="Password" onChange={handleInputPassword} />
      </Form.Item>

      {formMessage.message && (
        <Form.Item style={{ marginBottom: token.marginSM }}>
          <Alert
            message={formMessage.message}
            type={formMessage.type}
            showIcon
          />
        </Form.Item>
      )}
      <Form.Item>
        <StyledButton type="primary" htmlType="submit">
          {t("4QBwNXfHi4cI7j1q7aFw")}
        </StyledButton>
      </Form.Item>
    </StyledForm>
  );
};

export default StepEnterNewPassword;
