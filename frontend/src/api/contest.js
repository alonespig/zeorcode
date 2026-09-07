import request from '@/utils/request'

export const createContest = (data) => {
  return request({
    url: '/contest',
    method: 'post',
    data
  })
}

export const getContestDesc = (id) => {
  return request({
    url: `/contest/${id}/desc`,
    method: 'get',
  })
}

export const getContestProblemList = (id) => {
  return request({
    url: `/contest/${id}/problem`,
    method: 'get',
  })
}


export const getContestList = (params) => {
  return request({
    url: '/contest',
    method: 'get',
    params
  })
}

export const joinContest = (data) => {
  return request({
    url: `/contest/join`,
    method: 'post',
    data
  })
}


export const getContestDetail = (id) => {
  return request({
    url: `/contest/${id}`,
    method: 'get'
  })
}

export const getContestRank = (id) => {
  return request({
    url: `/contest/${id}/rank`,
    method: 'get'
  })
}

// 当前用户在该比赛的名次
export const getMyContestRank = (id) => {
  return request({
    url: `/contest/${id}/myrank`,
    method: 'get'
  })
}


export const getContestProblem = ({contestID, problemId}) => {
  return request({
    url: `/contest/${contestID}/problem/${problemId}`,
    method: 'get'
  })
}

export const getContestSubmissionInfo = (id) => {
  return request({
    url: `/contest/${id}/submit-info`,
    method: 'get'
  })
}

export const submit = (data) => {
  return request({
    url: `/contest/submit`,
    method: 'post',
    data
  })
}

export const getSubmission = (id, params) => {
  return request({
    url: `/contest/${id}/submission`,
    method: 'get',
    params,
  })
}

// 整场重判（超管）
export const rejudgeContest = (id) => {
  return request({
    url: `/admin/contest/${id}/rejudge`,
    method: 'post',
  })
}

// 某比赛某题全部重判（超管）
export const rejudgeContestProblem = (contestID, problemID) => {
  return request({
    url: `/admin/contest/${contestID}/problem/${problemID}/rejudge`,
    method: 'post',
  })
}

// 编辑比赛：读取可编辑信息
export const getContestInfo = (id) => {
  return request({
    url: `/contest/${id}/info`,
    method: 'get',
  })
}

// 编辑比赛：保存修改
export const updateContest = (id, data) => {
  return request({
    url: `/contest/${id}`,
    method: 'put',
    data,
  })
}
