import { http } from "@/utils/http";
import type { ApiResponse } from "@/api/auth";

export type ContestStatus = 0 | 1 | 2;

export interface ContestItem {
  id: number;
  name: string;
  description: string;
  coverUrl: string;
  startTime: string;
  endTime: string;
  type: string;
  status: ContestStatus;
  isRegistered: boolean;
  rated: boolean;
  needInviteCode: boolean;
  duration: number;
  participants: number;
  archived: boolean;
}

export interface ContestListResp {
  total: number;
  list: ContestItem[];
}

export interface ContestListParams {
  page: number;
  pageSize: number;
  keyword?: string;
  type?: number;
  status?: number;
  archived?: boolean;
}

export interface ContestProblemInput {
  problemId: string;
  color: string;
  score: number;
}

export interface CreateContestParams {
  name: string;
  contestDate: string;
  contestTime: string;
  description: string;
  coverUrl: string;
  duration: number;
  ruleType: number;
  problems: ContestProblemInput[];
  inviteCode: string;
  rated: boolean;
}

export interface UpdateContestParams {
  title: string;
  description: string;
  coverUrl: string;
  ruleType: number;
  startTime: number;
  endTime: number;
  problems: ContestProblemInput[];
  rated: boolean;
}

export interface ContestEditProblem {
  id: string;
  name: string;
  color: string;
  score: number;
}

export interface ContestEditInfo {
  title: string;
  description: string;
  coverUrl: string;
  ruleType: number;
  startTime: number;
  endTime: number;
  problems: ContestEditProblem[];
  rated: boolean;
}

export function getAdminContestList(params: ContestListParams) {
  return http.request<ApiResponse<ContestListResp>>("get", "/admin/contests", {
    params
  });
}

export function setContestArchived(id: number, archived: boolean) {
  return http.request<ApiResponse>("put", `/admin/contest/${id}/archive`, {
    data: { archived }
  });
}

export function createContest(data: CreateContestParams) {
  return http.request<ApiResponse<{ id: number }>>("post", "/contest", {
    data
  });
}

export function getContestInfo(id: number) {
  return http.request<ApiResponse<ContestEditInfo>>(
    "get",
    `/contest/${id}/info`
  );
}

export function updateContest(id: number, data: UpdateContestParams) {
  return http.request<ApiResponse<{ id: number }>>("put", `/contest/${id}`, {
    data
  });
}

export function rejudgeContest(id: number) {
  return http.request<ApiResponse<{ rejudged: number }>>(
    "post",
    `/admin/contest/${id}/rejudge`
  );
}

export function recomputeContest(id: number) {
  return http.request<ApiResponse>("post", `/admin/contest/${id}/recompute`);
}
