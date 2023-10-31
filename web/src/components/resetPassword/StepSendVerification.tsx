import { ResetPasswordSteps, StepComponentParams } from "@/constant";
import { StyledButton, StyledForm, StyledInput } from "@/pages/signUp/style";
import { Form, type AlertProps, theme, Alert } from "antd";
import { isValidateEmail } from "@/util";
import { sendVerificationCode } from '@/api/account';
import { useState } from "react";
import { useTranslation } from "react-i18next";
import useStore from "@/store";

const validateFormFailed = ({
  email,
}: { email?: string }): string => {
  if (email?.trim() === "") {
    return "email is required";
  }
  if (!isValidateEmail(email?.trim() || "")) {
    return "email address not valid";
  }

  return "";
};

// The first step to reset the password, and confirm the registered email number
const StepSendVerification = ({ setStep }: StepComponentParams) => {
  const { t } = useTranslation();
  const [form] = Form.useForm<{ email?: string }>();
  const [isLoading, setIsLoading] = useState(false);
  const [formMessage, setFormMessage] = useState<{
    type: AlertProps["type"];
    message: string;
  }>({
    type: "error",
    message: "",
  });
  const { token } = theme.useToken();
  const setResetInfo = useStore((state) => state.setResetInfo);

  const handleSendVerification = async () => {
    const params = form.getFieldsValue();
    const errorMessage = validateFormFailed(params);
    if (errorMessage) {
      setFormMessage({
        type: 'error',
        message: errorMessage,
      });
      return;
    }

    setIsLoading(true);

    const result = await sendVerificationCode({
      action: 'reset_password',
      email: params.email!,
    });

    setIsLoading(false);

    if (result.error) {
      setFormMessage({
        type: 'error',
        message: result.error.message || 'send verification code fail.',
      });
      return;
    }

    setResetInfo({
      email: params.email,
      verificationToken: result.data?.verification_token,
    });
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
        <StyledButton type="primary" htmlType="submit" loading={isLoading}>
          {t("9GIeWlOpzcbz7B4nDnnhz")}
        </StyledButton>
      </Form.Item>
    </StyledForm>
  );
};

export default StepSendVerification;
