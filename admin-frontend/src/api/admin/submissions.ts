import { http } from "@/utils/http";
import type { ApiResponse } from "@/api/auth";

export interface SubmissionItem {
  id: number;
  problemID: string;
  problemName: string;
  userID: number;
  userName: string;
  rating: number;
  language: string;
  result: number;
  timeUsed: number;
  memoryUsed: number;
  createdAt: string;
}

export interface SubmissionListResp {
  total: number;
  list: SubmissionItem[];
}

export interface SubmissionListParams {
  page: number;
  pageSize: number;
  status?: number;
  username?: string;
  problemID?: string;
}

export interface SubmissionCaseResult {
  id: number;
  status: number;
  time: number;
  memory: number;
}

export interface SubmissionDetail {
  user: {
    id: number;
    name: string;
    avatar: string;
    gender: number;
  };
  problem: {
    id: string;
    name: string;
    description: string;
    language: string;
    code: string;
    status: number;
    oj: string;
  };
  submission: {
    id: number;
    language: string;
    code: string;
    canViewCode: boolean;
    status: number;
    time: number;
    memory: number;
    compileOutput: string;
    createdAt: string;
  };
  caseResults: SubmissionCaseResult[];
}

export function getSubmissionList(params: SubmissionListParams) {
  return http.request<ApiResponse<SubmissionListResp>>("get", "/submission", {
    params
  });
}

export function getSubmissionDetail(id: number) {
  return http.request<ApiResponse<SubmissionDetail>>(
    "get",
    `/submission/${id}`
  );
}

export function rejudgeSubmission(id: number) {
  return http.request<ApiResponse<{ rejudged: number }>>(
    "post",
    `/admin/submission/${id}/rejudge`
  );
}
