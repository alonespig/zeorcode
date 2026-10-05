import { http } from "@/utils/http";
import type { ApiResponse } from "@/api/auth";

/** 算法标签（对应后端 dto.TagItem） */
export interface TagItem {
  id: number;
  name: string;
}

/** 标签列表响应（GET /api/tags） */
export interface TagListResp {
  tags: TagItem[];
}

/** 题库列表项（对应后端 dto.ProblemItemResp） */
export interface ProblemItem {
  /** 对外题号，如 P1001 */
  id: string;
  name: string;
  /** 当前用户做题状态，未登录时为 nil */
  status?: number;
  /** 1 简单 / 2 中等 / 3 困难 */
  difficulty: number;
  tags: TagItem[];
  acceptedCount: number;
  submitCount: number;
  createdAt: string;
  /** 隐藏题（仅管理员能在列表看到） */
  hidden: boolean;
  /** 空=本地题；非空=远程题 */
  oj: string;
}

/** 题库分页列表响应（GET /api/problems） */
export interface ProblemListResp {
  total: number;
  list: ProblemItem[];
}

/** 题库列表查询参数（对应后端 dto.ProblemListReq） */
export interface ProblemListParams {
  page: number;
  pageSize: number;
  q?: string;
  /** 逗号分隔标签 id（AND 交集） */
  tags?: string;
  difficulty?: number;
}

/** 题目简要信息（按题号解析名称时使用，对应后端 dto.ProblemDetailResp 的 id/name） */
export interface ProblemBrief {
  id: string;
  name: string;
}

/** 全量算法标签（供题单筛选与表单下拉） */
export function getTagList() {
  return http.request<ApiResponse<TagListResp>>("get", "/tags");
}

/** 分页查询题库（供题单选题弹窗） */
export function getProblemList(params: ProblemListParams) {
  return http.request<ApiResponse<ProblemListResp>>("get", "/problems", {
    params
  });
}

/** 按对外题号查询题目，用于输入题号后解析名称 */
export function getProblem(id: string) {
  return http.request<ApiResponse<ProblemBrief>>("get", `/problems/${id}`);
}
