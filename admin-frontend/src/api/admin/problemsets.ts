import { http } from "@/utils/http";
import type { ApiResponse } from "@/api/auth";
import type { TagItem } from "@/api/admin/problems";

/** 题单列表项（对应后端 dto.ProblemSetItemResp） */
export interface ProblemSetItem {
  id: number;
  title: string;
  tags: TagItem[];
  problemCount: number;
  /** 当前用户已通过题数，未登录为 nil；后台列表忽略 */
  solvedCount?: number;
  /** 0 公开 / 1 需邀请码 */
  visibility: number;
  /** 0 草稿 / 1 已发布 */
  published: number;
  author: string;
  updatedAt: string;
}

/** 题单列表响应（对应后端 dto.ProblemSetListResp） */
export interface ProblemSetListResp {
  total: number;
  list: ProblemSetItem[];
}

/** 题单列表查询参数（对应后端 dto.ProblemSetListReq） */
export interface ProblemSetListParams {
  page: number;
  pageSize: number;
  q?: string;
  /** 逗号分隔标签 id（AND 交集） */
  tags?: string;
  /** 0 公开 / 1 需邀请码；不传=全部 */
  visibility?: number;
}

/** 题单内的单题（对应后端 dto.ProblemSetProblemResp） */
export interface ProblemSetProblem {
  id: string;
  name: string;
  status?: number;
  difficulty: number;
  tags: TagItem[];
  acceptedCount: number;
  submitCount: number;
}

/** 题单详情（对应后端 dto.ProblemSetDetailResp） */
export interface ProblemSetDetailResp {
  id: number;
  title: string;
  description: string;
  tags: TagItem[];
  visibility: number;
  published: number;
  problemCount: number;
  solvedCount?: number;
  /** 邀请码题单未解锁时为 true，problems 为空；管理员恒为 false */
  locked: boolean;
  author: string;
  updatedAt: string;
  problems: ProblemSetProblem[];
}

/** 新建/编辑题单请求（对应后端 dto.SaveProblemSetReq） */
export interface SaveProblemSetParams {
  title: string;
  description: string;
  /** 0 草稿 / 1 已发布 */
  published: number;
  /** 0 公开 / 1 需邀请码 */
  visibility: number;
  /** 编辑时留空表示保留原邀请码 */
  inviteCode: string;
  tagIds: number[];
  /** 对外题号列表，顺序敏感 */
  problems: string[];
}

/** 新建题单响应（POST 返回 { id }） */
export interface SaveProblemSetResult {
  id: number;
}

/** 后台分页查询题单（含草稿） */
export function getAdminProblemSetList(params: ProblemSetListParams) {
  return http.request<ApiResponse<ProblemSetListResp>>(
    "get",
    "/admin/problemset",
    { params }
  );
}

/** 题单详情（管理员可见草稿与未解锁内容） */
export function getProblemSetDetail(id: number) {
  return http.request<ApiResponse<ProblemSetDetailResp>>(
    "get",
    `/problemset/${id}`
  );
}

/** 新建题单 */
export function createProblemSet(data: SaveProblemSetParams) {
  return http.request<ApiResponse<SaveProblemSetResult>>(
    "post",
    "/admin/problemset",
    { data }
  );
}

/** 编辑题单 */
export function updateProblemSet(id: number, data: SaveProblemSetParams) {
  return http.request<ApiResponse>("put", `/admin/problemset/${id}`, { data });
}

/** 删除题单 */
export function deleteProblemSet(id: number) {
  return http.request<ApiResponse>("delete", `/admin/problemset/${id}`);
}
