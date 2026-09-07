import request from '@/utils/request'

export const getAgentConversations = (params = {}) => request({
  url: '/agent/conversations',
  method: 'get',
  params: { page: 1, pageSize: 30, ...params },
})

export const createAgentConversation = (data = {}) => request({
  url: '/agent/conversations',
  method: 'post',
  data,
})

export const getAgentConversation = (id) => request({
  url: `/agent/conversations/${id}`,
  method: 'get',
})

export const updateAgentConversation = (id, data) => request({
  url: `/agent/conversations/${id}`,
  method: 'patch',
  data,
})
