import { http } from "@/utils/http";
import type { ApiResponse } from "@/api/auth";

export interface DashboardSummary {
  users: number;
  problems: number;
  todaySubmissions: number;
  pendingPosts: number;
  runningContests: number;
  pendingJudging: number;
}

export interface DashboardTrendItem {
  date: string;
  count: number;
}

export interface DashboardSubmission {
  id: number;
  userID: number;
  username: string;
  problemID: string;
  problemName: string;
  status: number;
  createdAt: string;
}

export interface DashboardOverview {
  summary: DashboardSummary;
  submissionTrend: DashboardTrendItem[];
  recentSubmissions: DashboardSubmission[];
}

export function getDashboardOverview() {
  return http.request<ApiResponse<DashboardOverview>>(
    "get",
    "/admin/dashboard"
  );
}
