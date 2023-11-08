import { useSearchParams } from "react-router-dom";
import NormalLayout from '@/components/common/layout/NormalLayout';
import { Wrapper } from './style';

import verificationSuccess from '@/assets/images/verification-success.png';
import verificationFail from '@/assets/images/verification-fail.png';

function VerificationSuccess() {
  const [query] = useSearchParams();

  const isSuccess = query.get('success') === '1';
  const isFail = query.get('success') === '0';
  const isExpired = query.get('expired') === '1';

  return (
    <NormalLayout logo>
      <Wrapper>
        {isSuccess && (
          <>
            <img width={220} src={verificationSuccess} />
            <span style={{ fontSize: 14 }}>Email verification successful.</span>
          </>
        )}
        {isFail && (
          <>
            <img width={220} src={verificationFail} />
            <span style={{ fontSize: 14 }}>Email verification failed, please try again later.</span>
          </>
        )}
        {isExpired && (
          <>
            <img width={220} src={verificationFail} />
            <span style={{ fontSize: 14 }}>This link has expired.</span>
          </>
        )}
      </Wrapper>
    </NormalLayout>
  )
}

export default VerificationSuccess;
