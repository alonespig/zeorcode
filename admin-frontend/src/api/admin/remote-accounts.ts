import { http } from "@/utils/http";
import type { ApiResponse } from "@/api/auth";

export type RemoteAuthType = "password" | "cookie";

/** 远程账号列表项（对应后端 dto.RemoteAccountItem） */
export interface RemoteAccountItem {
  id: number;
  oj: string;
  authType: RemoteAuthType;
  username: string;
  hasSecret: boolean;
  enabled: boolean;
  valid: boolean;
  busy: boolean;
  created_at: number;
}

export interface RemoteAccountListResp {
  list: RemoteAccountItem[];
}

export interface RemoteOJListResp {
  list: string[];
}

export interface CreateRemoteAccountParams {
  oj: string;
  authType: RemoteAuthType;
  username: string;
  secret: string;
  enabled: boolean;
}

export type UpdateRemoteAccountParams = Omit<CreateRemoteAccountParams, "oj">;

export function getRemoteAccounts() {
  return http.request<ApiResponse<RemoteAccountListResp>>(
    "get",
    "/admin/remote-account"
  );
}

export function getRemoteOJList() {
  return http.request<ApiResponse<RemoteOJListResp>>("get", "/remote/oj");
}

export function createRemoteAccount(data: CreateRemoteAccountParams) {
  return http.request<ApiResponse<{ id: number }>>(
    "post",
    "/admin/remote-account",
    { data }
  );
}

export function updateRemoteAccount(
  id: number,
  data: UpdateRemoteAccountParams
) {
  return http.request<ApiResponse>("put", `/admin/remote-account/${id}`, {
    data
  });
}

export function deleteRemoteAccount(id: number) {
  return http.request<ApiResponse>("delete", `/admin/remote-account/${id}`);
}
