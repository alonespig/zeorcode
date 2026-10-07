import { http } from "@/utils/http";
import { notifyAuthExpired } from "@/utils/session-events";
import type { ApiResponse } from "@/api/auth";

const AUTH_FAILURE_CODES = new Set([20001, 20002, 20003]);
const API_BASE = (import.meta.env.VITE_API_BASE_URL || "/api").replace(
  /\/$/,
  ""
);

export interface AgentConversationListParams {
  page: number;
  pageSize: number;
  q?: string;
}

export interface AgentConversationItem {
  id: number;
  title: string;
  status: number;
  updatedAt: string;
}

export interface AgentConversationList {
  total: number;
  list: AgentConversationItem[];
}

export interface AgentOption {
  label: string;
  value: string | number | boolean;
  description?: string;
  disabled?: boolean;
  recommended?: boolean;
}

export interface AgentTableColumn {
  key: string;
  label: string;
  width?: number;
  align?: "left" | "center" | "right";
  formatter?: string;
}

export interface AgentBlock {
  type: string;
  requestId?: number;
  title?: string;
  description?: string;
  required?: boolean;
  options?: AgentOption[];
  columns?: AgentTableColumn[];
  rows?: Record<string, unknown>[];
  payload?: Record<string, unknown>;
}

export interface AgentMessage {
  id: number | string;
  role: "user" | "assistant" | string;
  kind?: string;
  content: string;
  blocks: AgentBlock[];
  createdAt?: string;
}

export interface AgentConversationDetail {
  id: number;
  title: string;
  status: number;
  pendingRequestId?: number;
  pendingActionId?: number;
  messages: AgentMessage[];
}

export interface AgentTurnParams {
  type:
    | "message"
    | "interaction_response"
    | "action_approval"
    | "action_rejection";
  content?: string;
  requestId?: number;
  value?: unknown;
  actionId?: number;
  draftVersion?: number;
}

export interface AgentEvent {
  type: string;
  runId?: number;
  data?: unknown;
}

export function getAgentConversations(params: AgentConversationListParams) {
  return http.request<ApiResponse<AgentConversationList>>(
    "get",
    "/agent/conversations",
    { params }
  );
}

export function createAgentConversation(title = "新对话") {
  return http.request<ApiResponse<AgentConversationItem>>(
    "post",
    "/agent/conversations",
    { data: { title } }
  );
}

export function getAgentConversation(id: number) {
  return http.request<ApiResponse<AgentConversationDetail>>(
    "get",
    `/agent/conversations/${id}`
  );
}

export function updateAgentConversation(
  id: number,
  params: { title?: string; archived?: boolean }
) {
  return http.request<ApiResponse>("patch", `/agent/conversations/${id}`, {
    data: params
  });
}

export async function streamAgentTurn(
  conversationId: number,
  params: AgentTurnParams,
  signal: AbortSignal,
  onEvent: (event: AgentEvent) => void
) {
  const response = await fetch(
    `${API_BASE}/agent/conversations/${conversationId}/turns`,
    {
      method: "POST",
      credentials: "include",
      headers: {
        Accept: "text/event-stream",
        "Content-Type": "application/json",
        "X-Requested-With": "XMLHttpRequest"
      },
      body: JSON.stringify(params),
      signal
    }
  );
  const contentType = response.headers.get("content-type") || "";
  if (!contentType.includes("text/event-stream")) {
    const result = (await response.json().catch(() => null)) as {
      code?: number;
      msg?: string;
    } | null;
    if (result?.code !== undefined && AUTH_FAILURE_CODES.has(result.code)) {
      notifyAuthExpired();
    }
    throw new Error(result?.msg || "AI 助手暂时不可用");
  }
  if (!response.body) throw new Error("当前浏览器不支持流式响应");

  const reader = response.body.getReader();
  const decoder = new TextDecoder("utf-8");
  let buffer = "";
  const emitChunk = (chunk: string) => {
    let eventName = "";
    const dataLines: string[] = [];
    for (const line of chunk.split("\n")) {
      if (line.startsWith("event:")) eventName = line.slice(6).trim();
      if (line.startsWith("data:")) dataLines.push(line.slice(5).trimStart());
    }
    if (!dataLines.length) return;
    const event = JSON.parse(dataLines.join("\n")) as AgentEvent;
    if (eventName) event.type = eventName;
    onEvent(event);
  };

  while (true) {
    const { value, done } = await reader.read();
    buffer += decoder.decode(value || new Uint8Array(), { stream: !done });
    const chunks = buffer.replace(/\r\n/g, "\n").split("\n\n");
    buffer = chunks.pop() || "";
    chunks.forEach(emitChunk);
    if (done) break;
  }
  if (buffer.trim()) emitChunk(buffer);
}
