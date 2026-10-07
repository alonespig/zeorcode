import { http } from "@/utils/http";
import type { ApiResponse } from "@/api/auth";

/** 编程语言（对应后端 dto.LanguageItem） */
export interface LanguageItem {
  id: number;
  name: string;
  status: 0 | 1;
  sort: number;
  createdAt: string;
  updatedAt: string;
}

export interface SaveLanguageParams {
  name: string;
  status: 0 | 1;
  sort: number;
}

export function getAdminLanguages() {
  return http.request<ApiResponse<LanguageItem[]>>("get", "/admin/languages");
}

export function createLanguage(data: SaveLanguageParams) {
  return http.request<ApiResponse<LanguageItem>>("post", "/admin/languages", {
    data
  });
}

export function updateLanguage(id: number, data: SaveLanguageParams) {
  return http.request<ApiResponse>("put", `/admin/languages/${id}`, { data });
}

export function deleteLanguage(id: number) {
  return http.request<ApiResponse>("delete", `/admin/languages/${id}`);
}
