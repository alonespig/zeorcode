import { http } from "@/utils/http";
import type { ApiResponse } from "@/api/auth";

export interface PermissionItem {
  key: string;
  label: string;
}

export interface PermissionGroup {
  key: string;
  label: string;
  permissions: PermissionItem[];
}

export interface PermissionCatalog {
  groups: PermissionGroup[];
}

export interface UserPermissions {
  role: number;
  permissions: string[];
}

export function getPermissionCatalog() {
  return http.request<ApiResponse<PermissionCatalog>>(
    "get",
    "/admin/permissions/catalog"
  );
}

export function getUserPermissions(id: string | number) {
  return http.request<ApiResponse<UserPermissions>>(
    "get",
    `/admin/users/${id}/permissions`
  );
}

export function setUserPermissions(id: string | number, permissions: string[]) {
  return http.request<ApiResponse>("put", `/admin/users/${id}/permissions`, {
    data: { permissions }
  });
}
