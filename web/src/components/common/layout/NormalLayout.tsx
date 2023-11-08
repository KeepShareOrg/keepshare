import React from 'react';
import { NormalLayout as Layout, NormalHeader, LogoBox, LogoImg, NormalContent } from './style';

import LogoWithText from '@/assets/images/logo-with-text.png';

interface NormalLayoutProps extends React.PropsWithChildren {
  logo?: boolean;
}

function NormalLayout(props: NormalLayoutProps) {
  const { logo } = props;
  return (
    <Layout>
      <NormalHeader>
        {logo && (
          <LogoBox>
            <LogoImg src={LogoWithText} />
          </LogoBox>
        )}
      </NormalHeader>
      <NormalContent>{props.children}</NormalContent>
    </Layout>
  )
}

export default NormalLayout;
