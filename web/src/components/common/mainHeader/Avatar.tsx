import { Dropdown, Space, Typography } from "antd";
import type { MenuProps } from "antd";
import { AvatarImg, AvatarMenuItem } from "./style";
import Avatar1 from "@/assets/avatar/avatar1.png";
import { LogoutOutlined, UserOutlined } from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import { RoutePaths } from "@/router";
import useStore from "@/store";
import { useEffect, useState } from "react";
import { getUserInfo } from "@/api/account";

const { Text } = Typography;

const LogoutItem = () => {
  const navigate = useNavigate();
  const signOut = useStore((state) => state.signOut);

  const handleLogout = () => {
    signOut();
    navigate(RoutePaths.Login);
  };

  return (
    <AvatarMenuItem onClick={handleLogout}>
      <Space>
        <LogoutOutlined />
        <Text>Log out</Text>
      </Space>
    </AvatarMenuItem>
  );
};

const ManageAccountItem = () => {
  const navigate = useNavigate();

  const handleClick = () => navigate(RoutePaths.Settings);

  return (
    <AvatarMenuItem onClick={handleClick}>
      <Space>
        <UserOutlined />
        <Text>Manage Account</Text>
      </Space>
    </AvatarMenuItem>
  );
};

const getAvatarMenu = (email?: string): MenuProps["items"] => {
  return [
    {
      key: "0",
      label: <AvatarMenuItem>{email || "-"}</AvatarMenuItem>,
    },
    {
      type: "divider",
    },
    {
      key: "1",
      label: <ManageAccountItem />,
    },
    {
      key: "2",
      label: <LogoutItem />,
    },
  ];
};

const Avatar = () => {
  const [userInfo, setUserInfo] = useStore((state) => [
    state.userInfo,
    state.setUserInfo,
  ]);

  useEffect(() => {
    getUserInfo().then(({ data }) => {
      data && setUserInfo(data);
    });
  }, []);

  const [avatarMenu, setAvatarMenu] = useState<MenuProps["items"]>([]);

  useEffect(() => {
    const menu = getAvatarMenu(userInfo.email);
    setAvatarMenu(menu);
  }, [userInfo.email]);

  return (
    <>
      <Dropdown menu={{ items: avatarMenu }} trigger={["click"]}>
        <AvatarImg src={Avatar1} />
      </Dropdown>
    </>
  );
};

export default Avatar;
