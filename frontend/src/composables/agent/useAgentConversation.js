import { computed, onScopeDispose, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  createAgentConversation,
  getAgentConversation,
  getAgentConversations,
  updateAgentConversation,
} from '@/api/agent'
import { useUserStore } from '@/stores/user'

const API_BASE = (import.meta.env.VITE_API_BASE_URL || '/api').replace(/\/$/, '')

export function useAgentConversation() {
	const userStore = useUserStore()
  const conversations = ref([])
  const current = ref(null)
  const listLoading = ref(false)
  const detailLoading = ref(false)
  const running = ref(false)
  const streamingContent = ref('')
	let activeController = null

  const messages = computed(() => Array.isArray(current.value?.messages) ? current.value.messages : [])
  const pendingRequestId = computed(() => current.value?.pendingRequestId || 0)
  const pendingActionId = computed(() => current.value?.pendingActionId || 0)

  async function loadConversations() {
    listLoading.value = true
    try {
      const response = await getAgentConversations()
      conversations.value = response.data?.list || []
      return conversations.value
    } finally {
      listLoading.value = false
    }
  }

  async function loadConversation(id, { quiet = false } = {}) {
    if (!id) return
    if (!quiet) detailLoading.value = true
    try {
      const response = await getAgentConversation(id)
      current.value = response.data
    } finally {
      if (!quiet) detailLoading.value = false
    }
  }

  async function createConversation() {
    const response = await createAgentConversation({ title: '新对话' })
    await loadConversations()
    await loadConversation(response.data.id)
    return response.data.id
  }

  async function archiveConversation(id) {
    await updateAgentConversation(id, { archived: true })
    if (current.value?.id === id) current.value = null
    await loadConversations()
  }

  async function ensureConversation() {
    if (current.value?.id) return current.value.id
    return createConversation()
  }

  async function runTurn(payload) {
    if (running.value) return
    const conversationId = await ensureConversation()
    running.value = true
    streamingContent.value = ''
		activeController = new AbortController()
    try {
      const response = await fetch(`${API_BASE}/agent/conversations/${conversationId}/turns`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json', Accept: 'text/event-stream' },
        body: JSON.stringify(payload),
			signal: activeController.signal,
      })
      const contentType = response.headers.get('content-type') || ''
      if (!contentType.includes('text/event-stream')) {
        const result = await response.json().catch(() => null)
			if ([20001, 20002, 20003].includes(result?.code)) {
				userStore.localLogout()
				userStore.promptLogin()
			}
        throw new Error(result?.msg || 'AI 助手暂时不可用')
      }
      if (!response.body) throw new Error('浏览器不支持流式响应')

      const reader = response.body.getReader()
      const decoder = new TextDecoder('utf-8')
      let buffer = ''
      while (true) {
        const { value, done } = await reader.read()
        buffer += decoder.decode(value || new Uint8Array(), { stream: !done })
        const chunks = buffer.replace(/\r\n/g, '\n').split('\n\n')
        buffer = chunks.pop() || ''
        for (const chunk of chunks) handleSSEChunk(chunk)
        if (done) break
      }
      if (buffer.trim()) handleSSEChunk(buffer)
    } catch (error) {
			if (error?.name !== 'AbortError') ElMessage.error(error?.message || 'AI 助手暂时不可用')
    } finally {
			activeController = null
      running.value = false
      streamingContent.value = ''
      const refreshResults = await Promise.allSettled([
        loadConversation(conversationId, { quiet: true }),
        loadConversations(),
      ])
      const refreshError = refreshResults.find((result) => result.status === 'rejected')
      if (refreshError) console.error('[Agent] 刷新对话失败', refreshError.reason)
    }
  }

  function handleSSEChunk(chunk) {
    let eventName = ''
    const dataLines = []
    for (const line of chunk.split('\n')) {
      if (line.startsWith('event:')) eventName = line.slice(6).trim()
      if (line.startsWith('data:')) dataLines.push(line.slice(5).trimStart())
    }
    if (!dataLines.length) return
    const event = JSON.parse(dataLines.join('\n'))
    const type = eventName || event.type
    if (type === 'message.delta') {
      streamingContent.value += event.data?.content || ''
    } else if (type === 'run.failed') {
      throw new Error(event.data?.message || '本次处理失败')
    }
  }

  const sendMessage = (content) => runTurn({ type: 'message', content })
  const answerInteraction = (requestId, value) => runTurn({
    type: 'interaction_response', requestId, value,
  })
  const approveAction = (actionId, draftVersion) => runTurn({
    type: 'action_approval', actionId, draftVersion,
  })
  const rejectAction = (actionId, draftVersion) => runTurn({
    type: 'action_rejection', actionId, draftVersion,
  })

	onScopeDispose(() => activeController?.abort())

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
    loadConversations,
    loadConversation,
    createConversation,
    archiveConversation,
    sendMessage,
    answerInteraction,
    approveAction,
    rejectAction,
  }
}
