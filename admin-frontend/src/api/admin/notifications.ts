import { http } from "@/utils/http";
import type { ApiResponse } from "@/api/auth";

export interface BroadcastNotificationParams {
  title: string;
  content: string;
  link: string;
}

export function broadcastNotification(data: BroadcastNotificationParams) {
  return http.request<ApiResponse>("post", "/admin/notifications", { data });
}
