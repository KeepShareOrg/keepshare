import { ResetPasswordSteps, StepComponentParams } from "@/constant";
import { StyledButton, StyledForm, StyledInput } from "@/pages/signUp/style";
import { Form, type AlertProps, theme, Alert } from "antd";
import { useState } from "react";
import { useTranslation } from "react-i18next";

// The first step to reset the password, and confirm the registered email number
const StepSendVerification = ({ setStep }: StepComponentParams) => {
  const { t } = useTranslation();
  const [form] = Form.useForm<{ email?: string }>();
  const [formMessage, setFormMessage] = useState<{
    type: AlertProps["type"];
    message: string;
  }>({
    type: "error",
    message: "",
  });
  const { token } = theme.useToken();

  // TODO:
  console.log(setFormMessage);

  const handleSendVerification = () => {
    setStep(ResetPasswordSteps.ENTER_VERIFICATION_CODE);
  };

  return (
    <StyledForm
      form={form}
      layout="vertical"
      onFinish={handleSendVerification}
      validateTrigger={[]}
      autoComplete="off"
    >
      <Form.Item name="email" label="Email">
        <StyledInput placeholder="Email address" />
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
          {t("9GIeWlOpzcbz7B4nDnnhz")}
        </StyledButton>
      </Form.Item>
    </StyledForm>
  );
};

export default StepSendVerification;
