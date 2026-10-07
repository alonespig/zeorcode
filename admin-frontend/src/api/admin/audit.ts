import { http } from "@/utils/http";
import type { ApiResponse } from "@/api/auth";

export interface AuditLogItem {
  id: number;
  actorID: number;
  actorName: string;
  method: string;
  path: string;
  target: string;
  clientIP: string;
  success: boolean;
  code: number;
  createdAt: string;
}

export interface AuditLogListResp {
  total: number;
  list: AuditLogItem[];
}

export interface AuditLogListParams {
  page: number;
  pageSize: number;
  q?: string;
  method?: string;
  success?: boolean;
}

export function getAuditLogs(params: AuditLogListParams) {
  return http.request<ApiResponse<AuditLogListResp>>(
    "get",
    "/admin/audit-logs",
    { params }
  );
}
