import { http } from "@/utils/http";

export interface ApiResponse<T = unknown> {
  code: number;
  msg: string;
  data?: T;
}

export interface LoginParams {
  username: string;
  password: string;
  captchaId: string;
  captchaCode: string;
}

export interface AdminUser {
  id: number;
  username: string;
  role: number;
  avatar?: string;
}

export interface LoginResult {
  user: AdminUser;
  avatar: string;
}

export interface SessionResult {
  authenticated: boolean;
  user?: AdminUser;
  avatar?: string;
}

export interface CaptchaResult {
  id: string;
  image: string;
}

export function getCaptcha() {
  return http.request<ApiResponse<CaptchaResult>>("get", "/captcha");
}

export function login(params: LoginParams) {
  return http.request<ApiResponse<LoginResult>>("post", "/login", {
    data: params
  });
}

export function getSession() {
  return http.request<ApiResponse<SessionResult>>("get", "/session");
}

export function logout() {
  return http.request<ApiResponse>("post", "/logout");
}
