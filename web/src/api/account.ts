import { axiosWrapper } from ".";

export interface SignUpParams {
  email: string;
  password_hash: string;
  captcha_token: string;
}
export interface SignUpResponse {
  ok: boolean;
  access_token: string;
  refresh_token: string;
}
// register keep share account by email
export const signUp = (params: SignUpParams) => {
  return axiosWrapper<SignUpResponse>({
    method: "post",
    url: "/session/sign_up",
    data: params,
  });
};

export interface SignInParams {
  email: string;
  password_hash: string;
  captcha_token?: string;
}
export interface SignInResponse {
  ok: boolean;
  access_token: string;
  refresh_token: string;
}
// login to keep share
export const signIn = (params: SignInParams) => {
  return axiosWrapper<SignInResponse>({
    method: "post",
    url: "/session/sign_in",
    data: params,
  });
};

export interface RefreshTokenResponse {
  access_token: string;
  refresh_token: string;
}
// refresh token by refresh_token
export const refreshToken = (refreshToken: string) => {
  return axiosWrapper<RefreshTokenResponse>({
    method: "post",
    url: "/session/token",
    data: {
      refresh_token: refreshToken,
    },
  });
};

export interface GetUserInfoResponse {
  user_id: string;
  channel_id: string;
  email: string;
  username: string;
}
// get user keep share account info
export const getUserInfo = () => {
  return axiosWrapper<GetUserInfoResponse>({
    method: "get",
    url: "/session/me",
  });
};


export interface SendVerifyEmailResponse {
  success: boolean;
}
// send verification email
export const sendVerifyEmail = () => {
  return axiosWrapper<SendVerifyEmailResponse>({
    method: "post",
    url: "/api/verification",
  });
};

export interface ChangeAccountEmailParams {
  new_email: string;
  password_hash: string;
}
export interface ChangeAccountEmailResponse {
  success: boolean;
}
export const changeAccountEmail = (params: ChangeAccountEmailParams) => {
  return axiosWrapper<ChangeAccountEmailResponse>({
    method: "post",
    url: "/api/change_email",
    data: params,
  });
};

export interface ChangeAccountPasswordParams {
  password_hash: string;
  new_password_hash: string;
}
export interface ChangeAccountPasswordResponse {
  success: boolean;
}
export const changeAccountPassword = (params: ChangeAccountPasswordParams) => {
  return axiosWrapper<ChangeAccountPasswordResponse>({
    method: "post",
    url: "/api/change_password",
    data: params,
  });
};

export interface SendVerificationCodeParams {
  action: 'reset_password',
  email: string;
}
export interface SendVerificationCodeResponse {
  success: boolean;
  verification_token: string;
}
// send verification code
export const sendVerificationCode = (params: SendVerificationCodeParams) => {
  return axiosWrapper<SendVerificationCodeResponse>({
    method: "post",
    url: "/api/send_verification_code",
    data: params,
  });
};

export interface ResetAccountPasswordParams {
  email: string;
  password_hash: string;
  action: 'reset_password';
  verification_token: string;
  verification_code: string;
}
export interface ResetAccountPasswordResponse {
  success: boolean;
}
// reset account password
export const resetAccountPassword = (params: ResetAccountPasswordParams) => {
  return axiosWrapper<ResetAccountPasswordResponse>({
    method: "post",
    url: "/api/reset_password",
    data: params,
  });
};