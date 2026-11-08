import request from '@/utils/request'

export const getPostList = (params) => {
  return request({
    url: '/posts',
    method: 'get',
    params,
  })
}

export const getPost = (id) => {
  return request({
    url: `/posts/${id}`,
    method: 'get',
  })
}

export const createPost = (data) => {
  return request({
    url: '/posts',
    method: 'post',
    data,
  })
}

export const updatePost = (id, data) => {
  return request({
    url: `/posts/${id}`,
    method: 'put',
    data,
  })
}

export const deletePost = (id) => {
  return request({
    url: `/posts/${id}`,
    method: 'delete',
  })
}

export const togglePostLike = (id) => {
  return request({
    url: `/posts/${id}/like`,
    method: 'post',
  })
}

export const getPostComments = (id, params) => {
  return request({
    url: `/posts/${id}/comments`,
    method: 'get',
    params,
  })
}

export const createPostComment = (id, data) => {
  return request({
    url: `/posts/${id}/comments`,
    method: 'post',
    data,
  })
}

export const deletePostComment = (id) => {
  return request({
    url: `/comments/${id}`,
    method: 'delete',
  })
}

export const toggleCommentLike = (id) => {
  return request({
    url: `/comments/${id}/like`,
    method: 'post',
  })
}
