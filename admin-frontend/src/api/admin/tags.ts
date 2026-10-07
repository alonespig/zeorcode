import { http } from "@/utils/http";
import type { ApiResponse } from "@/api/auth";

/** 后台标签列表项（对应后端 dto.AdminTagItem） */
export interface AdminTagItem {
  id: number;
  name: string;
  problemCount: number;
  problemSetCount: number;
}

/** 后台标签列表响应 */
export interface AdminTagListResp {
  tags: AdminTagItem[];
}

/** 标签写入参数（对应后端 dto.SaveTagReq） */
export interface SaveTagParams {
  name: string;
}

/** 标签写入响应（对应后端 dto.TagItem） */
export interface TagItem {
  id: number;
  name: string;
}

export function getAdminTagList() {
  return http.request<ApiResponse<AdminTagListResp>>("get", "/admin/tags");
}

export function createTag(data: SaveTagParams) {
  return http.request<ApiResponse<TagItem>>("post", "/admin/tags", { data });
}

export function updateTag(id: number, data: SaveTagParams) {
  return http.request<ApiResponse<TagItem>>("put", `/admin/tags/${id}`, {
    data
  });
}

export function deleteTag(id: number) {
  return http.request<ApiResponse>("delete", `/admin/tags/${id}`);
}
