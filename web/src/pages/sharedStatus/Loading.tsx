import { Space, Typography } from "antd";
import { useState, type ReactNode, useEffect } from "react";
import { useTranslation } from "react-i18next";
import LoadingAPng from "@/assets/images/keepshare-loading.png";

const { Text } = Typography;

interface ComponentInterface {
  children: ReactNode;
}
const Loading = ({ children }: ComponentInterface) => {
  const { t } = useTranslation();

  const [countDown, setCountDown] = useState(10);

  useEffect(() => {
    if (countDown === 0) return;

    const timerId = setTimeout(() => {
      setCountDown(countDown - 1);
    }, 1000);

    return () => clearTimeout(timerId);
  }, [countDown]);

  return (
    <Space
      direction="vertical"
      align="center"
      style={{
        position: "fixed",
        top: "50%",
        left: "50%",
        transform: "translate(-50%, -50%)",
      }}
    >
      <Space>{children}</Space>
      <Space direction="horizontal" style={{ marginTop: "20px" }}>
        <Space>
          <img src={LoadingAPng} alt="loading" width={100} />
          <Text style={{ display: "inline-block", maxWidth: "200px" }}>
            {t("tHxXtk0qRYf6Kh4qNcuHh")}({countDown})
          </Text>
        </Space>
      </Space>
    </Space>
  );
};

export default Loading;
