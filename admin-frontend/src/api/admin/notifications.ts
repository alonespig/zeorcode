import { http } from "@/utils/http";
import type { ApiResponse } from "@/api/auth";

export interface BroadcastNotificationParams {
  title: string;
  content: string;
  link: string;
}

export interface BroadcastHistoryItem {
  id: number;
  actorID: number;
  actorName: string;
  title: string;
  content: string;
  link: string;
  recipientCount: number;
  createdAt: string;
}

export interface BroadcastHistoryResp {
  total: number;
  list: BroadcastHistoryItem[];
}

export function broadcastNotification(data: BroadcastNotificationParams) {
  return http.request<ApiResponse>("post", "/admin/notifications", { data });
}

export function getBroadcastHistory(params: {
  page: number;
  pageSize: number;
}) {
  return http.request<ApiResponse<BroadcastHistoryResp>>(
    "get",
    "/admin/notifications",
    { params }
  );
}
