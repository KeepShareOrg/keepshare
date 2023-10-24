import {
  ProfileOutlined,
  CloudUploadOutlined,
  ShareAltOutlined,
  SettingOutlined,
  ClusterOutlined,
} from "@ant-design/icons";
import type { MenuProps } from "antd";
import { Menu } from "antd";
import PikPakLogo from "@/assets/images/PikPak.png";
import MegaLogo from "@/assets/images/Mega.png";
import { useLocation, useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import { RoutePaths } from "@/router";
import useStore from "@/store";

type MenuItem = Required<MenuProps>["items"][number];

const getMenuItems: () => MenuItem[] = () => [
  {
    key: RoutePaths.AutoShare,
    label: "Auto Share",
    icon: <ProfileOutlined />,
  },
  {
    key: RoutePaths.LinkToShare,
    label: "Link to Share",
    icon: <CloudUploadOutlined />,
  },
  {
    key: RoutePaths.SharedLinks,
    label: "Shared Links",
    icon: <ShareAltOutlined />,
  },
  {
    key: RoutePaths.Settings,
    label: "Settings",
    icon: <SettingOutlined />,
  },
  {
    key: "management",
    label: "Host Management",
    icon: <ClusterOutlined />,
    children: [
      {
        key: RoutePaths.PikPak,
        label: "PikPak",
        icon: (
          <img src={PikPakLogo} style={{ width: "14px", height: "14px" }} />
        ),
      },
      {
        key: RoutePaths.Mega,
        label: "Mega",
        icon: <img src={MegaLogo} style={{ width: "14px", height: "14px" }} />,
      },
    ],
  },
];

type MenuClickEventHandler = MenuProps["onClick"];
const MainMenu = () => {
  const navigate = useNavigate();
  const { pathname } = useLocation();
  const [currentMenu, setCurrentMenu] = useState([pathname]);
  const [isMobile, setShowMenuDrawer] = useStore((state) => [
    state.isMobile,
    state.setShowMenuDrawer,
  ]);

  const handleItemClick: MenuClickEventHandler = ({ key, keyPath }) => {
    setCurrentMenu(keyPath as RoutePaths[]);
    navigate(key);
    isMobile && setShowMenuDrawer(false);
  };

  useEffect(() => setCurrentMenu([pathname as RoutePaths]), [pathname]);

  return (
    <Menu
      rootClassName="keep-share-custom-menu-style"
      selectedKeys={currentMenu}
      mode="inline"
      items={getMenuItems()}
      onClick={handleItemClick}
    />
  );
};

export default MainMenu;
