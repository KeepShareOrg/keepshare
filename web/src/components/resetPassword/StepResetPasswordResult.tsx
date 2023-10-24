import { StyledButton, StyledForm } from "@/pages/signUp/style";
import { RoutePaths } from "@/router";
import { theme, type AlertProps, Form, Alert } from "antd";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

// Reset password result
const StepResetPasswordResult = () => {
  const [formMessage, setFormMessage] = useState<{
    type: AlertProps["type"];
    message: string;
  }>({
    type: "success",
    message: "Your password has been reset. Please log in to your account now.",
  });
  const { t } = useTranslation();
  const { token } = theme.useToken();

  // TODO:
  console.log(setFormMessage);

  const navigate = useNavigate();
  const handleBackToLogin = () => navigate(RoutePaths.Login);

  return (
    <StyledForm onFinish={handleBackToLogin}>
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
          {t("gwO6taePcRfEuC8Tx3")}
        </StyledButton>
      </Form.Item>
    </StyledForm>
  );
};

export default StepResetPasswordResult;
