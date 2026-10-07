import { http } from "@/utils/http";
import type { ApiResponse } from "@/api/auth";
import type { AxiosResponse } from "axios";

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
  /** latest 时按创建顺序倒序 */
  order?: "latest";
}

export interface ProblemSample {
  input: string;
  output: string;
  explain: string;
}

export interface SaveProblemParams {
  displayId: string;
  name: string;
  difficulty: number;
  timeLimit: number;
  memoryLimit: number;
  description: string;
  inputFormat: string;
  outputFormat: string;
  hint: string;
  samples: ProblemSample[];
  tagsID: number[];
  hidden: boolean;
}

export interface CreateProblemParams extends SaveProblemParams {
  oj: string;
  remoteProblemId: string;
}

/** 编辑回填结构（后端 dto.ProblemModel 的 tagsID 实际为标签对象数组） */
export interface ProblemForEdit {
  id: string;
  name: string;
  difficulty: number;
  timeLimit: number;
  memoryLimit: number;
  description: string;
  inputFormat: string;
  outputFormat: string;
  hint: string;
  samples: ProblemSample[];
  tagsID: TagItem[];
  hidden: boolean;
}

export interface RemoteProblem {
  oj: string;
  remoteProblemId: string;
  title: string;
  timeLimit: number;
  memoryLimit: number;
  description: string;
  inputFormat: string;
  outputFormat: string;
  samples: ProblemSample[];
  hint: string;
  source: string;
}

export interface TestDataFile {
  name: string;
  size: number;
}

export interface TestDataPreview {
  name: string;
  size: number;
  content: string;
  truncated: boolean;
}

export interface TestDataImportResult {
  count: number;
  missing: string[];
}

/** 题目简要信息（按题号解析名称时使用，对应后端 dto.ProblemDetailResp 的 id/name） */
export interface ProblemBrief {
  id: string;
  name: string;
  oj?: string;
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

export function createProblem(data: CreateProblemParams) {
  return http.request<ApiResponse<{ id: string }>>("post", "/problems", {
    data
  });
}

export function updateProblem(id: string, data: SaveProblemParams) {
  return http.request<ApiResponse<{ id: string }>>("put", `/problems/${id}`, {
    data
  });
}

export function getProblemForEdit(id: string) {
  return http.request<ApiResponse<ProblemForEdit>>(
    "get",
    `/problems/${id}/edit`
  );
}

export function getRemoteProblem(params: { oj: string; pid: string }) {
  return http.request<ApiResponse<RemoteProblem>>("get", "/remote/problem", {
    params
  });
}

export function getTestDataFiles(id: string) {
  return http.request<ApiResponse<TestDataFile[]>>(
    "get",
    `/problems/${id}/testdata`
  );
}

export function getTestDataContent(id: string, file: string) {
  return http.request<ApiResponse<TestDataPreview>>(
    "get",
    `/problems/${id}/testdata/content`,
    { params: { file } }
  );
}

export function uploadTestDataFile(id: string, data: FormData) {
  return http.request<ApiResponse<TestDataImportResult | undefined>>(
    "post",
    `/problems/${id}/testdata`,
    { data, timeout: 120_000 }
  );
}

export function replaceTestDataZip(id: string, data: FormData) {
  return http.request<ApiResponse<TestDataImportResult>>(
    "post",
    `/problems/${id}/testdata/zip`,
    { data, timeout: 120_000 }
  );
}

export function deleteTestDataFiles(id: string, files: string[]) {
  return http.request<ApiResponse<{ failed?: string[] }>>(
    "delete",
    `/problems/${id}/testdata`,
    { data: { files } }
  );
}

export async function downloadTestDataFiles(
  id: string,
  files: string[]
): Promise<Blob> {
  const response = await http.request<AxiosResponse<Blob>>(
    "get",
    `/problems/${id}/testdata/download`,
    {
      params: { files: files.join(",") },
      responseType: "blob"
    }
  );
  return response.data;
}
