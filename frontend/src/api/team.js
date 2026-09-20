import request from '@/utils/request'

// ===== 团队 =====

// 团队列表。params: { page, pageSize, q, mine, visibility }
export const getTeamList = (params) => {
  return request({ url: '/team', method: 'get', params })
}

export const getTeamDetail = (id) => {
  return request({ url: `/team/${id}`, method: 'get' })
}

export const createTeam = (data) => {
  return request({ url: '/team', method: 'post', data })
}

export const updateTeam = (id, data) => {
  return request({ url: `/team/${id}`, method: 'put', data })
}

export const deleteTeam = (id) => {
  return request({ url: `/team/${id}`, method: 'delete' })
}

// 加入团队（非公开团队需带 inviteCode）
export const joinTeam = (id, inviteCode) => {
  return request({ url: `/team/${id}/join`, method: 'post', data: { inviteCode } })
}

export const quitTeam = (id) => {
  return request({ url: `/team/${id}/quit`, method: 'post' })
}

// ===== 成员 =====

export const getTeamMembers = (id) => {
  return request({ url: `/team/${id}/member`, method: 'get' })
}

// 设置/取消团队管理员（role: 0 成员 / 1 管理员）
export const setMemberRole = (teamId, uid, role) => {
  return request({ url: `/team/${teamId}/member/${uid}/role`, method: 'put', data: { role } })
}

export const removeMember = (teamId, uid) => {
  return request({ url: `/team/${teamId}/member/${uid}`, method: 'delete' })
}

export const importTeamStudents = (teamId, file) => {
  const data = new FormData()
  data.append('file', file)
  return request({
    url: `/team/${teamId}/member/import`,
    method: 'post',
    data,
    timeout: 120000,
  })
}

export const importTeamStudentsManual = (teamId, text) => {
  return request({
    url: `/team/${teamId}/member/import/manual`,
    method: 'post',
    data: { text },
    timeout: 120000,
  })
}

// ===== 作业 =====

export const getHomeworkList = (teamId, params) => {
  return request({ url: `/team/${teamId}/homework`, method: 'get', params })
}

export const getHomeworkDetail = (id) => {
  return request({ url: `/homework/${id}`, method: 'get' })
}

export const getHomeworkProblem = (id, problemId) => {
  return request({ url: `/homework/${id}/problem/${problemId}`, method: 'get' })
}

// 布置作业（teamId 用于 POST /team/:id/homework）
export const createHomework = (teamId, data) => {
  return request({ url: `/team/${teamId}/homework`, method: 'post', data })
}

// 修改作业（方案 B：仅布置者本人可改）
export const updateHomework = (id, data) => {
  return request({ url: `/homework/${id}`, method: 'put', data })
}

export const deleteHomework = (id) => {
  return request({ url: `/homework/${id}`, method: 'delete' })
}

export const submitHomework = (id, data) => {
  return request({ url: `/homework/${id}/submit`, method: 'post', data })
}

export const getHomeworkRank = (id) => {
  return request({ url: `/homework/${id}/rank`, method: 'get' })
}

export const downloadHomeworkRank = (id) => {
  return request({ url: `/homework/${id}/rank/export`, method: 'get', responseType: 'blob' })
}

export const getHomeworkSubmissions = (id, params) => {
  return request({ url: `/homework/${id}/submission`, method: 'get', params })
}

export const getHomeworkSubmissionDetail = (homeworkId, submissionId) => {
  return request({ url: `/homework/${homeworkId}/submission/${submissionId}`, method: 'get' })
}
