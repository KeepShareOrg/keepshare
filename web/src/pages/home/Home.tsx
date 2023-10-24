import MainLayout from "@/layout/MainLayout";
import MobileMainLayout from "@/layout/MainLayout.m";
import { RoutePaths } from "@/router";
import useStore from "@/store";
import { useEffect } from "react";
import { useNavigate } from "react-router-dom";

const Home = () => {
  const navigate = useNavigate();
  useEffect(() => {
    /^\/console\/?$/i.test(window.location.pathname) &&
      navigate(RoutePaths.AutoShare);
  }, []);

  const isMobile = useStore((state) => state.isMobile);

  return isMobile ? <MobileMainLayout /> : <MainLayout />;
};

export default Home;
