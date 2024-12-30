import { ResetPasswordSteps, StepComponentParams } from "@/constant";
import { StyledButton, StyledForm, StyledInput } from "@/pages/signUp/style";
import { theme, type AlertProps, Form, Alert } from "antd";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import useStore from "@/store";

const validateFormFailed = ({
  verificationCode,
}: {
  verificationCode?: string;
}): string => {
  if (verificationCode?.trim() === "") {
    return "VerificationCode is null";
  }

  if (verificationCode && verificationCode.trim().length < 6) {
    return "Length is at least 6";
  }

  return "";
};

// Reset the second step of the password and verify the received mailbox verification code
const StepEnterVerificationCode = ({ setStep }: StepComponentParams) => {
  const { t } = useTranslation();
  const [form] = Form.useForm();
  const [formMessage, setFormMessage] = useState<{
    type: AlertProps["type"];
    message: string;
  }>({
    type: "error",
    message: "",
  });
  const { token } = theme.useToken();
  const email = useStore((state) => state.resetInfo.email);
  const setResetInfo = useStore((state) => state.setResetInfo);

  const handleVerify = () => {
    const params = form.getFieldsValue();
    const errorMessage = validateFormFailed(params);
    if (errorMessage) {
      setFormMessage({
        type: "error",
        message: errorMessage,
      });
      return;
    }

    setResetInfo({
      verificationCode: params.verificationCode,
    });
    setStep(ResetPasswordSteps.ENTER_NEW_PASSWORD);
  };

  return (
    <StyledForm
      form={form}
      layout="vertical"
      onFinish={handleVerify}
      validateTrigger={[]}
      autoComplete="off"
    >
      <Form.Item
        name="verificationCode"
        label={`An email with a verification code was just sent to ${email}`}
      >
        <StyledInput placeholder="Verification code" maxLength={6} />
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
          {t("e6jFx3MkyYdhhMq_0Hkxx")}
        </StyledButton>
      </Form.Item>
    </StyledForm>
  );
};

export default StepEnterVerificationCode;
