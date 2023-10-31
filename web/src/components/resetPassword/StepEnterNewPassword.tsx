import { ResetPasswordSteps, StepComponentParams } from "@/constant";
import { StyledButton, StyledForm, PasswordInput } from "@/pages/signUp/style";
import { type AlertProps, theme, Form, Alert } from "antd";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { resetAccountPassword } from '@/api/account';
import useStore from "@/store";
import { calcPasswordHash } from "@/util";

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

  const resetInfo = useStore(state => state.resetInfo);
  const setResetInfo = useStore(state => state.setResetInfo);
  const { token } = theme.useToken();

  const [formMessage, setFormMessage] = useState<{
    type: AlertProps["type"];
    message: string;
  }>({
    type: "error",
    message: "",
  });
  const [isLoading, setIsLoading] = useState(false);

  const handleResetPassword = async () => {

    const params = form.getFieldsValue();

    setIsLoading(true);

    const result = await resetAccountPassword({
      email: resetInfo.email,
      password_hash: calcPasswordHash(params.password!),
      action: 'reset_password',
      verification_token: resetInfo.verificationToken,
      verification_code: resetInfo.verificationCode,
    });

    setIsLoading(false);

    if (result.error) {
      return setFormMessage({ type: "error", message: result.error.message || 'reset fail.' });
    }

    setResetInfo({
      email: '',
      verificationCode: '',
      verificationToken: '',
    });
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
      onValuesChange={handleInputPassword}
      validateTrigger={[]}
      autoComplete="off"
    >
      <Form.Item
        name="password"
        label="Please set a new password"
        style={{ marginBottom: token.marginSM }}
      >
        <PasswordInput placeholder="Password" />
      </Form.Item>
      <Form.Item name="confirmPassword">
        <PasswordInput 
        placeholder="Confirm your password" />
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
          {t("4QBwNXfHi4cI7j1q7aFw")}
        </StyledButton>
      </Form.Item>
    </StyledForm>
  );
};

export default StepEnterNewPassword;
