import Logo from "@/assets/images/logo.png";
import { Wrapper, Avatar, Title } from "./style";
import { theme } from "antd";

interface ComponentInterface {
  title: string;
}
const AccountBanner = ({ title }: ComponentInterface) => {
  const { token } = theme.useToken();
  return (
    <Wrapper>
      <Avatar src={Logo} alt="avatar" />
      <Title style={{ color: token.colorText }}>{title}</Title>
    </Wrapper>
  );
};

export default AccountBanner;
