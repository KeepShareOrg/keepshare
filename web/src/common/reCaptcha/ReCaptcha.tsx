import { CAPTCHA_SITE_KEY } from "@/config";
import useStore from "@/store";
import { Spin } from "antd";
import { createRef, useState } from "react";
import GoogleReCaptcha from "react-google-recaptcha";

interface ComponentInterface {
  handleToken: (token: string) => void;
}

const ReCaptcha = ({ handleToken }: ComponentInterface) => {
  const [isLoading, setLoading] = useState(true);
  const reCaptchaRef = createRef<GoogleReCaptcha>();
  const themeMode = useStore((state) => state.themeMode);

  const handleCaptchaChange = (token: string | null) =>
    token && handleToken(token);

  const asyncScriptOnLoad = () => setLoading(false);

  return (
    <>
      {isLoading && <Spin />}
      <GoogleReCaptcha
        theme={themeMode === "dark" ? "dark" : "light"}
        ref={reCaptchaRef}
        sitekey={CAPTCHA_SITE_KEY}
        onChange={handleCaptchaChange}
        asyncScriptOnLoad={asyncScriptOnLoad}
      />
    </>
  );
};

export default ReCaptcha;
