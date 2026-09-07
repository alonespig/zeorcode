import request from '@/utils/request'

export const createProblem = (data) => {
  return request({
    url: '/problems',
    method: 'post',
    data
  })
}

export const getTags = () => {
  return request({
    url: '/tags',
    method: 'get'
  })
}

export const getProblemList = (params) => {
  return request({
    url: '/problems',
    method: 'get',
    params
  })
}

export const getProblem = (id) => {
  return request({
    url: `/problems/${id}`,
    method: 'get'
  })
}

export const updateProblem = (id, data) => {
  return request({
    url: `/problems/${id}`,
    method: 'put',
    data
  })
}

export const getProblemInfo = (id) => {
  return request({
    url: `/problems/${id}/edit`,
    method: 'get'
  })
}

export const fileList = (id) => {
  return request({
    url: `/problems/${id}/testdata`,
    method: 'get'
  })
}

export const getTestDataContent = (id, name) => {
  return request({
    url: `/problems/${id}/testdata/content`,
    method: 'get',
    params: { file: name }
  })
}

export const fileDelete = (id, data) => {
  return request({
    url: `/problems/${id}/testdata`,
    method: 'delete',
    data
  })
}

// 上传测试点压缩包(zip)：后端只提取 N.in/N.out，原子替换该题测试数据
export const uploadTestcaseZip = (id, formData, config = {}) => {
  return request({
    url: `/problems/${id}/testdata/zip`,
    method: 'post',
    data: formData,
    ...config
  })
}
