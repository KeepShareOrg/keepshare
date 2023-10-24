import { Form, message, theme, Alert } from "antd";
import AccountBanner from "@/components/account/accountBanner/AccountBanner";
import DefaultLayout from "@/layout/DefaultLayout";
import {
  StyledForm,
  StyledButton,
  SignUpTips,
  Wrapper,
  StyledInput,
  StyledItem,
  PasswordInput,
} from "./style";
import { Link, useNavigate } from "react-router-dom";
import ReCAPTCHA from "@/common/reCaptcha/ReCaptcha";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { signUp } from "@/api/account";
import { calcPasswordHash, isValidateEmail } from "@/util";
import { RoutePaths } from "@/router";

interface FieldType {
  email?: string;
  password?: string;
  confirmPassword?: string;
  captchaToken?: string;
}
type ErrorMessage = string;
const validateFromFailed = ({
  email,
  password,
  confirmPassword,
  captchaToken,
}: FieldType): ErrorMessage => {
  if (!email || email?.trim() === "") {
    return "email is required";
  }
  if (!password || password?.trim() === "") {
    return "password is required";
  }
  if (!confirmPassword || confirmPassword?.trim() === "") {
    return "confirm password is required";
  }
  if (!captchaToken || captchaToken?.trim() === "") {
    return "please verify captcha first";
  }
  if (!isValidateEmail(email?.trim() || "")) {
    return "email address not valid";
  }
  if (password !== confirmPassword) {
    return "The confirm password and password must match";
  }

  return "";
};

const SignUp = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm<FieldType>();
  const [shouldVerify] = useState(true);
  const [errorMessage, setErrorMessage] = useState("");

  const { token } = theme.useToken();
  const { t } = useTranslation();

  const fillReCaptchaToken = (token: string) => {
    form.setFieldValue("captchaToken", token);
  };

  const handleSignUp = async () => {
    try {
      const params = form.getFieldsValue();
      const errorMessage = validateFromFailed(params);
      if (errorMessage !== "") {
        return setErrorMessage(errorMessage);
      }
      setErrorMessage("");

      console.log("validate result: ", errorMessage, params);

      const { email, password, captchaToken } = params;
      const { data, error } = await signUp({
        email: email!,
        password_hash: calcPasswordHash(password!),
        captcha_token: captchaToken!,
      });

      console.log("data: ", data, "error: ", error);
      if (data?.ok) {
        message.success("sign up success!");
      } else {
        message.error("sign up error!");
      }

      navigate(RoutePaths.AutoShare);
    } catch (err) {
      console.error("handle login error: ", err);
    }
  };

  return (
    <DefaultLayout>
      <Wrapper>
        <AccountBanner title={t("efYgLz42PuSnMxL9wG9A")} />
        <StyledForm
          form={form}
          layout="vertical"
          onFinish={handleSignUp}
          validateTrigger={[]}
          autoComplete="off"
        >
          <Form.Item
            name="email"
            label="Email - only used to retrieve password"
          >
            <StyledInput placeholder="Email address" />
          </Form.Item>
          <StyledItem name="password" label="Password">
            <PasswordInput placeholder="Password" />
          </StyledItem>
          <StyledItem name="confirmPassword" label="Confirm Password">
            <PasswordInput placeholder="Repeat Password" />
          </StyledItem>
          {shouldVerify && (
            <Form.Item name="captchaToken">
              <ReCAPTCHA handleToken={fillReCaptchaToken} />
            </Form.Item>
          )}

          {errorMessage && (
            <Form.Item style={{ marginBottom: token.marginSM }}>
              <Alert message={errorMessage} type="error" showIcon />
            </Form.Item>
          )}
          <Form.Item style={{ marginBottom: token.marginSM }}>
            <StyledButton type="primary" htmlType="submit">
              Sign up
            </StyledButton>
          </Form.Item>
          <SignUpTips>
            <Form.Item noStyle> Already have an account? </Form.Item>
            <Link style={{ color: token.colorPrimary }} to={RoutePaths.Login}>
              Sign in
            </Link>
          </SignUpTips>
        </StyledForm>
      </Wrapper>
    </DefaultLayout>
  );
};

export default SignUp;
