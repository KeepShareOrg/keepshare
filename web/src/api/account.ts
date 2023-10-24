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
