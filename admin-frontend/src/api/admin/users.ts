import type { AxiosResponse } from "axios";
import { http } from "@/utils/http";
import type { ApiResponse } from "@/api/auth";

/** 管理员视角下的用户简要信息（对应后端 dto.AdminUserItem） */
export interface AdminUserItem {
  /** 对外用户号（UID） */
  id: number;
  username: string;
  /** 学号 */
  studentNo: string;
  /** 姓名 */
  realName: string;
  email: string;
  /** 角色：0 普通用户 / 1 管理员 */
  role: number;
  /** 状态：0 正常 / 1 封禁 */
  status: number;
  signature: string;
  /** 注册时间（秒级 Unix 时间戳） */
  createdAt: number;
}

/** 用户列表响应（对应后端 dto.AdminUserListResp） */
export interface AdminUserListResp {
  total: number;
  list: AdminUserItem[];
}

/** 用户列表分页参数（对应后端 dto.PageForm） */
export interface AdminUserListParams {
  page: number;
  pageSize: number;
  q?: string;
  role?: number;
  status?: number;
}

/** 封禁/解封用户请求（对应后端 dto.SetUserStatusReq） */
export interface SetUserStatusParams {
  /** 0 正常 / 1 封禁 */
  status: number;
}

/** 修改用户角色请求（对应后端 dto.SetUserRoleReq） */
export interface SetUserRoleParams {
  /** 0 普通用户 / 1 管理员 */
  role: number;
}

/** 批量导入用户结果（对应后端 dto.BatchCreateUsersResp） */
export interface BatchCreateUsersResult {
  /** 新建账号数量 */
  created: number;
}

/** 分页查询用户列表 */
export function getAdminUserList(params: AdminUserListParams) {
  return http.request<ApiResponse<AdminUserListResp>>("get", "/admin/users", {
    params
  });
}

/** 下载用户导入模板，返回 Blob（携带 HttpOnly Cookie） */
export async function downloadUserImportTemplate(): Promise<Blob> {
  const response = await http.request<AxiosResponse<Blob>>(
    "get",
    "/admin/users/import/template",
    { responseType: "blob" }
  );
  return response.data;
}

/** 批量导入用户，multipart 字段名为 file（携带 HttpOnly Cookie） */
export function importAdminUsers(file: File) {
  const formData = new FormData();
  formData.append("file", file);
  return http.request<ApiResponse<BatchCreateUsersResult>>(
    "post",
    "/admin/users/import",
    { data: formData, timeout: 120000 }
  );
}

/** 封禁/解封用户（0 正常 / 1 封禁） */
export function setUserStatus(id: number, params: SetUserStatusParams) {
  return http.request<ApiResponse>("put", `/admin/users/${id}/status`, {
    data: params
  });
}

/** 修改用户角色（0 普通用户 / 1 管理员） */
export function setUserRole(id: number, params: SetUserRoleParams) {
  return http.request<ApiResponse>("put", `/admin/users/${id}/role`, {
    data: params
  });
}
