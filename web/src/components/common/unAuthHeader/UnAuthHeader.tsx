import { Button, theme } from "antd";
import { StyledLink, NavWrapper } from "./style";
import { RoutePaths } from "@/router";

const UnAuthHeader = () => {
  const { token } = theme.useToken();

  return (
    <NavWrapper>
      <Button
        type="link"
        style={{
          fontSize: token.fontSizeLG,
          color: token.colorText,
        }}
        href={window.origin}
      >
        Home
      </Button>
      <StyledLink
        color={token.colorText}
        to={"https://github.com/KeepShareOrg/keepshare"}
      >
        Github
      </StyledLink>
      <StyledLink color={token.colorText} to={RoutePaths.Donation}>
        Donation
      </StyledLink>
    </NavWrapper>
  );
};

export default UnAuthHeader;
