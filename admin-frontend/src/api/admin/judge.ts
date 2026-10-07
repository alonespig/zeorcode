import { http } from "@/utils/http";
import type { ApiResponse } from "@/api/auth";

/** 单台评测机的实时探活结果（对应 judge.ProbeResult） */
export interface JudgeNodeStatus {
  url: string;
  online: boolean;
  latencyMs: number;
  version?: string;
  error?: string;
}

export interface JudgeStatusResp {
  list: JudgeNodeStatus[];
}

export function getJudgeStatus() {
  return http.request<ApiResponse<JudgeStatusResp>>(
    "get",
    "/admin/judge/status"
  );
}
