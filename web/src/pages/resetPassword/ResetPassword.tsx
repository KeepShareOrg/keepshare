import AccountBanner from "@/components/account/accountBanner/AccountBanner";
import DefaultLayout from "@/layout/DefaultLayout";
import { Wrapper } from "./style";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { ResetPasswordSteps } from "@/constant";
import StepSendVerification from "@/components/resetPassword/StepSendVerification";
import StepEnterVerificationCode from "@/components/resetPassword/StepEnterVerificationCode";
import StepEnterNewPassword from "@/components/resetPassword/StepEnterNewPassword";
import StepResetPasswordResult from "@/components/resetPassword/StepResetPasswordResult";

const ResetPassword = () => {
  const { t } = useTranslation();
  const [currentStep, setStep] = useState<ResetPasswordSteps>(
    ResetPasswordSteps.SEND_VERIFICATION_CODE,
  );

  return (
    <DefaultLayout>
      <Wrapper>
        <AccountBanner
          title={
            currentStep === ResetPasswordSteps.RESET_PASSWORD_RESULT
              ? "Congratulations!"
              : t("qA5YtlBvXJtD8qralNokn")
          }
        />
        {currentStep === ResetPasswordSteps.SEND_VERIFICATION_CODE && (
          <StepSendVerification setStep={setStep} />
        )}
        {currentStep === ResetPasswordSteps.ENTER_VERIFICATION_CODE && (
          <StepEnterVerificationCode setStep={setStep} />
        )}
        {currentStep === ResetPasswordSteps.ENTER_NEW_PASSWORD && (
          <StepEnterNewPassword setStep={setStep} />
        )}
        {currentStep === ResetPasswordSteps.RESET_PASSWORD_RESULT && (
          <StepResetPasswordResult />
        )}
      </Wrapper>
    </DefaultLayout>
  );
};

export default ResetPassword;
