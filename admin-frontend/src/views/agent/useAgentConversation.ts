import { computed, onScopeDispose, ref } from "vue";
import { message } from "@/utils/message";
import {
  createAgentConversation,
  getAgentConversation,
  getAgentConversations,
  streamAgentTurn,
  updateAgentConversation,
  type AgentConversationDetail,
  type AgentConversationItem,
  type AgentEvent,
  type AgentMessage,
  type AgentTurnParams
} from "@/api/admin/agent";

export function useAgentConversation() {
  const conversations = ref<AgentConversationItem[]>([]);
  const current = ref<AgentConversationDetail>();
  const listLoading = ref(false);
  const detailLoading = ref(false);
  const running = ref(false);
  const streamingContent = ref("");
  const toolLabel = ref("");
  const selectedConversationId = ref(0);
  let activeController: AbortController | undefined;
  let detailSequence = 0;

  const messages = computed(() => current.value?.messages || []);
  const pendingRequestId = computed(() => current.value?.pendingRequestId || 0);
  const pendingActionId = computed(() => current.value?.pendingActionId || 0);

  async function loadConversations(keyword = "") {
    listLoading.value = true;
    try {
      const response = await getAgentConversations({
        page: 1,
        pageSize: 50,
        q: keyword || undefined
      });
      conversations.value = response.data?.list || [];
      return conversations.value;
    } finally {
      listLoading.value = false;
    }
  }

  async function loadConversation(id: number, quiet = false) {
    const sequence = ++detailSequence;
    if (!quiet) detailLoading.value = true;
    try {
      const response = await getAgentConversation(id);
      if (sequence === detailSequence && response.data)
        current.value = response.data;
    } finally {
      if (!quiet && sequence === detailSequence) detailLoading.value = false;
    }
  }

  async function createConversation() {
    stopTurn();
    const response = await createAgentConversation();
    if (!response.data) throw new Error("创建对话失败");
    selectedConversationId.value = response.data.id;
    await loadConversations();
    await loadConversation(response.data.id);
  }

  async function archiveConversation(id: number) {
    stopTurn();
    await updateAgentConversation(id, { archived: true });
    if (current.value?.id === id) current.value = undefined;
    const list = await loadConversations();
    selectedConversationId.value = list[0]?.id || 0;
    if (list.length) await loadConversation(list[0].id);
  }

  function stopTurn() {
    activeController?.abort();
    activeController = undefined;
  }

  async function selectConversation(id: number) {
    if (id === current.value?.id) return;
    selectedConversationId.value = id;
    stopTurn();
    streamingContent.value = "";
    await loadConversation(id);
  }

  function handleEvent(event: AgentEvent) {
    if (event.type === "message.delta") {
      const data = event.data as { content?: string } | undefined;
      streamingContent.value += data?.content || "";
      return;
    }
    if (event.type === "tool.started") {
      const data = event.data as { label?: string } | undefined;
      toolLabel.value = data?.label || "正在处理";
      return;
    }
    if (event.type === "run.failed") {
      const data = event.data as { message?: string } | undefined;
      throw new Error(data?.message || "本次处理失败");
    }
  }

  async function ensureConversation() {
    if (current.value?.id) return current.value.id;
    const response = await createAgentConversation();
    if (!response.data) throw new Error("创建对话失败");
    selectedConversationId.value = response.data.id;
    await loadConversations();
    await loadConversation(response.data.id);
    return response.data.id;
  }

  async function runTurn(params: AgentTurnParams) {
    if (running.value) return;
    const conversationId = await ensureConversation();
    const optimisticMessage: AgentMessage | undefined =
      params.type === "message" && params.content
        ? {
            id: `pending-${Date.now()}`,
            role: "user",
            content: params.content,
            blocks: []
          }
        : undefined;
    if (optimisticMessage && current.value?.id === conversationId) {
      current.value.messages.push(optimisticMessage);
    }

    running.value = true;
    streamingContent.value = "";
    toolLabel.value = "正在思考";
    const controller = new AbortController();
    activeController = controller;
    try {
      await streamAgentTurn(
        conversationId,
        params,
        controller.signal,
        handleEvent
      );
    } catch (error) {
      if (!(error instanceof DOMException && error.name === "AbortError")) {
        message(error instanceof Error ? error.message : "AI 助手暂时不可用", {
          type: "error"
        });
      }
    } finally {
      if (activeController === controller) activeController = undefined;
      running.value = false;
      streamingContent.value = "";
      toolLabel.value = "";
      const tasks: Promise<unknown>[] = [loadConversations()];
      if (selectedConversationId.value === conversationId) {
        tasks.push(loadConversation(conversationId, true));
      }
      const results = await Promise.allSettled(tasks);
      const failed = results.find(result => result.status === "rejected");
      if (failed) console.error("[Agent] 刷新对话失败", failed.reason);
    }
  }

  onScopeDispose(stopTurn);

  return {
    conversations,
    current,
    messages,
    pendingRequestId,
    pendingActionId,
    listLoading,
    detailLoading,
    running,
    streamingContent,
    toolLabel,
    loadConversations,
    selectConversation,
    createConversation,
    archiveConversation,
    sendMessage: (content: string) => runTurn({ type: "message", content }),
    answerInteraction: (requestId: number, value: unknown) =>
      runTurn({ type: "interaction_response", requestId, value }),
    approveAction: (actionId: number, draftVersion: number) =>
      runTurn({ type: "action_approval", actionId, draftVersion }),
    rejectAction: (actionId: number, draftVersion: number) =>
      runTurn({ type: "action_rejection", actionId, draftVersion })
  };
}
