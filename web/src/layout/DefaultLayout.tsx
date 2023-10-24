import React from "react";
import { Background, ContentWrapper } from "./style";
import UnAuthHeader from "@/components/common/unAuthHeader/UnAuthHeader";

interface ComponentProps {
  children: React.ReactNode;
}
const DefaultLayout = ({ children }: ComponentProps) => {
  return (
    <Background>
      <UnAuthHeader />
      <ContentWrapper>{children}</ContentWrapper>
    </Background>
  );
};

export default DefaultLayout;
