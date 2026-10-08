import { http } from "@/utils/http";
import type { ApiResponse } from "@/api/auth";

/** 帖子作者简要信息（对应后端 dto.PostUser） */
export interface PostUser {
  id: number;
  username: string;
  avatar: string;
  rating: number;
}

/** 关联题目简要信息（对应后端 dto.ProblemSimple） */
export interface PostProblem {
  id: string;
  name: string;
}

/** 帖子列表项（对应后端 dto.PostItem） */
export interface PostItem {
  id: number;
  category: string;
  /** 0 待审核 / 1 通过 / 2 拒绝 */
  reviewStatus: number;
  rejectReason?: string;
  /** 对外题号 */
  problemID: string;
  problem?: PostProblem;
  title: string;
  summary: string;
  user: PostUser;
  likeCount: number;
  viewCount: number;
  commentCount: number;
  createdAt: string;
  updatedAt: string;
}

/** 帖子分页列表响应（对应后端 dto.PostListResp） */
export interface PostListResp {
  total: number;
  list: PostItem[];
}

/** 帖子详情（对应后端 dto.PostDetailResp，含正文） */
export interface PostDetailResp extends PostItem {
  content: string;
  isLiked: boolean;
}

/** 审核帖子请求（对应后端 dto.ReviewPostReq） */
export interface ReviewPostParams {
  /** 1 通过 / 2 拒绝 */
  status: 1 | 2;
  /** 拒绝理由，仅 status=2 时携带 */
  reason?: string;
}

/** 后台分页查询待审核帖子（GET /admin/posts/pending） */
export function getPendingPosts(params: { page: number; pageSize: number }) {
  return http.request<ApiResponse<PostListResp>>(
    "get",
    "/admin/posts/pending",
    { params }
  );
}

/** 帖子详情（GET /posts/:id，含正文） */
export function getPostDetail(id: number) {
  return http.request<ApiResponse<PostDetailResp>>("get", `/posts/${id}`);
}

/** 审核帖子（PUT /admin/posts/:id/review） */
export function reviewPost(id: number, status: 1 | 2, reason?: string) {
  return http.request<ApiResponse>("put", `/admin/posts/${id}/review`, {
    data: { status, reason } satisfies ReviewPostParams
  });
}
