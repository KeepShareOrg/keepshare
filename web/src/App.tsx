import { RouterProvider } from "react-router-dom";
import router from "@/router";
import { ConfigProvider } from "antd";
import lightTheme from "@/common/style/lightTheme";
import darkTheme from "@/common/style/darkTheme";
import useStore from "@/store";
import "@/locale";
import { useEffect } from "react";
import { CustomLightStyle, CustomDarkStyle } from "@/common/style/globalStyle";
import { getRememberToken } from "./util";

const App = () => {
  const [themeMode, setThemeMode, setKeepLogin, setIsMobile] = useStore(
    (state) => [
      state.themeMode,
      state.setThemeMode,
      state.setKeepLogin,
      state.setIsMobile,
    ],
  );

  useEffect(() => {
    const isDarkMode =
      window.matchMedia &&
      window.matchMedia("(prefers-color-scheme: dark)").matches;

    setThemeMode(isDarkMode ? "dark" : "light");

    const isKeepLogin = getRememberToken();
    setKeepLogin(isKeepLogin);
  }, []);

  useEffect(() => {
    const resize = () => setIsMobile(window.innerWidth < 768);
    resize();
    window.addEventListener("resize", resize);

    return () => window.removeEventListener("resize", resize);
  }, []);

  return (
    <>
      {themeMode === "dark" ? <CustomDarkStyle /> : <CustomLightStyle />}
      <ConfigProvider theme={themeMode === "dark" ? darkTheme : lightTheme}>
        <RouterProvider router={router} />
      </ConfigProvider>
    </>
  );
};

export default App;
