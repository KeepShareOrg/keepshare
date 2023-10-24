import { Form, Checkbox, theme, message, Alert } from "antd";
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
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { signIn } from "@/api/account";
import { calcPasswordHash, isValidateEmail, setRememberToken } from "@/util";
import { RoutePaths } from "@/router";
import useStore from "@/store";

interface FieldType {
  email?: string;
  password?: string;
  remember?: string;
  captchaToken?: string;
}
type ErrorMessage = string;
const validateFormFailed = ({
  email,
  password,
  captchaToken,
}: FieldType): ErrorMessage => {
  if (email?.trim() === "") {
    return "email is required";
  }
  if (password?.trim() === "") {
    return "password is required";
  }
  if (captchaToken?.trim() === "") {
    return "please verify captcha first";
  }
  if (!isValidateEmail(email?.trim() || "")) {
    return "email address not valid";
  }

  return "";
};

const Login = () => {
  const [form] = Form.useForm<FieldType>();
  const [shouldVerify] = useState(true);

  const { token } = theme.useToken();
  const { t } = useTranslation();

  let redirectPath = RoutePaths.Home;
  useEffect(() => {
    const routeParams = new URLSearchParams(window.location.search);
    const fromRoute = routeParams.get("from") || RoutePaths.Home;
    redirectPath = fromRoute as RoutePaths;
  }, []);

  const fillReCaptchaToken = (token: string) => {
    form.setFieldValue("captchaToken", token);
  };

  const keepLogin = useStore((state) => state.keepLogin);
  useEffect(() => {
    form.setFieldValue("remember", keepLogin);
  }, [keepLogin]);

  const [errorMessage, setErrorMessage] = useState("");
  const navigate = useNavigate();
  const handleLogin = async () => {
    try {
      const params = form.getFieldsValue();
      const errorMessage = validateFormFailed(params);
      if (errorMessage !== "") {
        setErrorMessage(errorMessage);
        return;
      }

      const { email, password, captchaToken, remember } = params;

      setRememberToken(remember ? "true" : "false");

      const { error } = await signIn({
        email: email!,
        password_hash: calcPasswordHash(password!),
        captcha_token: captchaToken,
      });

      if (error !== null) {
        message.error(error?.message || "error");
        return;
      }

      message.success("login success!");

      navigate(redirectPath);
    } catch (err) {
      console.error("handle login error: ", err);
    }
  };

  return (
    <DefaultLayout>
      <Wrapper>
        <AccountBanner title={t("nhs6biEWvwDk16pMvAd78")} />
        <StyledForm
          form={form}
          layout="vertical"
          onFinish={handleLogin}
          validateTrigger={[]}
          autoComplete="off"
        >
          <Form.Item name="email" label="Email">
            <StyledInput placeholder="Email address" />
          </Form.Item>
          <StyledItem name="password" label="Password">
            <PasswordInput placeholder="Password" />
          </StyledItem>
          {shouldVerify && (
            <Form.Item
              name="captchaToken"
              style={{ marginBottom: token.marginSM }}
            >
              <ReCAPTCHA handleToken={fillReCaptchaToken} />
            </Form.Item>
          )}
          <Form.Item>
            <Form.Item<FieldType>
              name="remember"
              valuePropName="checked"
              noStyle
            >
              <Checkbox>Keep me logged in</Checkbox>
            </Form.Item>

            <Link
              style={{ color: token.colorPrimary }}
              to={RoutePaths.ResetPassword}
            >
              Forgot password?
            </Link>
          </Form.Item>

          {errorMessage && (
            <Form.Item style={{ marginBottom: token.marginSM }}>
              <Alert message={errorMessage} type="error" showIcon />
            </Form.Item>
          )}
          <Form.Item style={{ marginBottom: token.marginSM }}>
            <StyledButton type="primary" htmlType="submit">
              Log in
            </StyledButton>
          </Form.Item>
          <SignUpTips>
            <Form.Item noStyle> Do not have an account? </Form.Item>
            <Link style={{ color: token.colorPrimary }} to={RoutePaths.SignUp}>
              Sign up
            </Link>
          </SignUpTips>
        </StyledForm>
      </Wrapper>
    </DefaultLayout>
  );
};

export default Login;
