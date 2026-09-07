package service

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"zoj/internal/common/errcode"
	"zoj/internal/dto"
	"zoj/internal/model"
	"zoj/internal/repository"

	"gorm.io/gorm"
)

const (
	postCategorySolution = "solution"
	postHelpSolution     = "help"
	defaultPostPage      = 1
	defaultPostPageSize  = 10
	maxPostPageSize      = 50
)

type PostService struct {
	repo        *repository.PostRepo
	userRepo    *repository.UserRepo
	problemRepo *repository.ProblemRepo
	notify      *NotificationService
}

func NewPostService(repo *repository.PostRepo, userRepo *repository.UserRepo, problemRepo *repository.ProblemRepo, notify *NotificationService) *PostService {
	return &PostService{repo: repo, userRepo: userRepo, problemRepo: problemRepo, notify: notify}
}

func postLookupError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.ErrPostNotFound
	}
	return errcode.ErrDatabase.Wrap(err)
}

func commentLookupError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.ErrCommentNotFound
	}
	return errcode.ErrDatabase.Wrap(err)
}

func (s *PostService) ListPosts(ctx context.Context, req *dto.PostListReq, viewerPK int64) (*dto.PostListResp, error) {
	normalizePostPage(req)
	// 按题目筛选时 req.ProblemID 是对外题号，解析成内部主键；解析不到直接返回空
	var problemPK int64
	if req.ProblemID != "" {
		pk, err := s.problemRepo.ResolveID(ctx, req.ProblemID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errcode.ErrDatabase.Wrap(err)
			}
			return &dto.PostListResp{Total: 0, List: []dto.PostItem{}}, nil
		}
		problemPK = pk
	}
	// 按用户筛选时 req.UserID 是对外用户号，解析成内部主键
	userPK := req.UserID
	if req.UserID > 0 {
		pk, err := s.userRepo.ResolveIDByUID(ctx, req.UserID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errcode.ErrDatabase.Wrap(err)
			}
			return &dto.PostListResp{Total: 0, List: []dto.PostItem{}}, nil
		}
		userPK = pk
	}
	// 公开列表只显示"已通过"；查看自己的帖子时(userID 过滤==本人)则显示自己全部(含待审/被拒)
	var reviewFilter *int
	if !(userPK != 0 && userPK == viewerPK) {
		approved := model.PostReviewApproved
		reviewFilter = &approved
	}
	posts, total, err := s.repo.List(ctx, req.Category, problemPK, userPK,
		strings.TrimSpace(req.Keyword), req.Page, req.PageSize, req.Sort == "hot", reviewFilter)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	items, err := s.buildPostItems(ctx, posts)
	if err != nil {
		return nil, err
	}
	return &dto.PostListResp{
		Total: total,
		List:  items,
	}, nil
}

// AdminListPending 管理员审核队列：列出待审核帖子（不受可见性限制）。
func (s *PostService) AdminListPending(ctx context.Context, page, pageSize int) (*dto.PostListResp, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	pending := model.PostReviewPending
	posts, total, err := s.repo.List(ctx, "all", 0, 0, "", page, pageSize, false, &pending)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	items, err := s.buildPostItems(ctx, posts)
	if err != nil {
		return nil, err
	}
	return &dto.PostListResp{Total: total, List: items}, nil
}

// ReviewPost 管理员通过/拒绝帖子。status 只接受 通过(1)/拒绝(2)；拒绝可带理由。
func (s *PostService) ReviewPost(ctx context.Context, id int64, status int, reason string) error {
	if status != model.PostReviewApproved && status != model.PostReviewRejected {
		return errcode.ErrInvalidParams.WithMsg("非法审核状态")
	}
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return postLookupError(err)
	}
	if status == model.PostReviewApproved {
		reason = "" // 通过则清空理由
	}
	if err := s.repo.Update(ctx, id, map[string]any{
		"review_status": status,
		"reject_reason": reason,
	}); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

func (s *PostService) CreatePost(ctx context.Context, req *dto.CreatePostReq, userID int64, role int) (*dto.CreatePostResp, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, errcode.ErrInvalidParams.WithMsg("标题不能为空")
	}
	if req.Category == "announcement" && role != 1 {
		return nil, errcode.ErrOperationDenied.WithMsg("只有管理员可以发布公告")
	}
	problemPK, err := s.resolvePostProblem(ctx, req.Category, req.ProblemID)
	if err != nil {
		return nil, err
	}
	summary := strings.TrimSpace(req.Summary)
	if summary == "" {
		summary = makePostSummary(req.Content)
	}
	// 管理员发的、以及公告，自动通过；其余普通用户帖子进待审
	reviewStatus := model.PostReviewPending
	if role == 1 || req.Category == "announcement" {
		reviewStatus = model.PostReviewApproved
	}
	post := &model.Post{
		UserID:       userID,
		Category:     req.Category,
		ProblemID:    problemPK,
		Title:        title,
		Content:      req.Content,
		Summary:      summary,
		ReviewStatus: reviewStatus,
	}
	if err := s.repo.Create(ctx, post); err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	return &dto.CreatePostResp{ID: post.ID, ReviewStatus: reviewStatus}, nil
}

func (s *PostService) UpdatePost(ctx context.Context, id int64, req *dto.UpdatePostReq, userID int64, role int) error {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return postLookupError(err)
	}
	if !canManagePost(post.UserID, userID, role) {
		return errcode.ErrOperationDenied
	}
	if req.Category == "announcement" && role != 1 {
		return errcode.ErrOperationDenied.WithMsg("只有管理员可以发布公告")
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return errcode.ErrInvalidParams.WithMsg("标题不能为空")
	}
	problemPK, err := s.resolvePostProblem(ctx, req.Category, req.ProblemID)
	if err != nil {
		return err
	}
	summary := strings.TrimSpace(req.Summary)
	if summary == "" {
		summary = makePostSummary(req.Content)
	}
	// 管理员编辑保持通过；普通用户编辑后回到待审并清空拒绝理由（需重新审核）
	reviewStatus := model.PostReviewApproved
	if role != 1 && req.Category != "announcement" {
		reviewStatus = model.PostReviewPending
	}
	err = s.repo.Update(ctx, id, map[string]any{
		"category":      req.Category,
		"problem_id":    problemPK,
		"title":         title,
		"content":       req.Content,
		"summary":       summary,
		"review_status": reviewStatus,
		"reject_reason": "",
	})
	if err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

func (s *PostService) DeletePost(ctx context.Context, id int64, userID int64, role int) error {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return postLookupError(err)
	}
	if !canManagePost(post.UserID, userID, role) {
		return errcode.ErrOperationDenied
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

func (s *PostService) GetPost(ctx context.Context, id int64, userID int64, role int) (*dto.PostDetailResp, error) {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, postLookupError(err)
	}
	// 未过审的帖子仅作者本人或管理员可见（对外一律当作不存在）
	if post.ReviewStatus != model.PostReviewApproved && post.UserID != userID && role != 1 {
		return nil, errcode.ErrPostNotFound
	}
	// 只有已通过的帖子才计浏览量（作者预览待审稿不刷数）
	if post.ReviewStatus == model.PostReviewApproved {
		if err := s.repo.IncrViewCount(ctx, id); err != nil {
			return nil, errcode.ErrDatabase.Wrap(err)
		}
		post.ViewCount++
	}

	items, err := s.buildPostItems(ctx, []model.Post{*post})
	if err != nil {
		return nil, err
	}
	item := items[0]
	resp := &dto.PostDetailResp{
		PostItem: item,
		Content:  post.Content,
	}
	if userID > 0 {
		liked, err := s.repo.HasLiked(ctx, id, userID)
		if err != nil {
			return nil, errcode.ErrDatabase.Wrap(err)
		}
		resp.IsLiked = liked
	}
	return resp, nil
}

func (s *PostService) ToggleLike(ctx context.Context, postID, userID int64) (*dto.TogglePostLikeResp, error) {
	post, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		return nil, postLookupError(err)
	}
	liked, err := s.repo.ToggleLike(ctx, postID, userID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	if liked { // 仅点赞（非取消）时通知帖子作者
		s.notify.Notify(ctx, post.UserID, userID, "like", post.Title, "", "/blog/"+strconv.FormatInt(postID, 10), "post", postID)
	}
	return &dto.TogglePostLikeResp{Liked: liked}, nil
}

// ToggleCommentLike 点赞/取消点赞评论；点赞时通知评论作者。
func (s *PostService) ToggleCommentLike(ctx context.Context, commentID, userID int64) (*dto.TogglePostLikeResp, error) {
	comment, err := s.repo.GetComment(ctx, commentID)
	if err != nil {
		return nil, commentLookupError(err)
	}
	liked, err := s.repo.ToggleCommentLike(ctx, commentID, userID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	if liked {
		s.notify.Notify(ctx, comment.UserID, userID, "like", "", "赞了你的评论",
			"/blog/"+strconv.FormatInt(comment.PostID, 10), "comment", commentID)
	}
	return &dto.TogglePostLikeResp{Liked: liked}, nil
}

// ListComments 返回单层楼中楼：顶层评论按时间序，每条顶层评论挂它的全部回复。
// 回复若回复的是另一条回复（而非直接回复楼主），ReplyTo 填被回复者，前端显示 @某人。
// Total 为评论总数（顶层 + 回复）。
func (s *PostService) ListComments(ctx context.Context, postID, userID int64) (*dto.PostCommentListResp, error) {
	if _, err := s.repo.GetByID(ctx, postID); err != nil {
		return nil, postLookupError(err)
	}
	comments, err := s.repo.ListComments(ctx, postID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	likedSet, _ := s.repo.LikedCommentIDs(ctx, postID, userID)
	roots, err := s.buildCommentTree(ctx, comments, likedSet)
	if err != nil {
		return nil, err
	}
	return &dto.PostCommentListResp{
		Total: int64(len(comments)),
		List:  roots,
	}, nil
}

func (s *PostService) CreateComment(ctx context.Context, postID, userID int64, req *dto.CreatePostCommentReq) error {
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return errcode.ErrInvalidParams.WithMsg("评论内容不能为空")
	}
	post, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		return postLookupError(err)
	}
	// 由「被回复的评论」推出 root_id 和 reply_user_id：
	//   回复一级评论(楼主) → root=该评论, 不@(reply_user_id=0)
	//   回复二级评论        → root=它所属的一级评论, @该二级评论作者
	var rootID, replyUserID, parentAuthor int64
	if req.ParentID > 0 {
		parent, err := s.repo.GetComment(ctx, req.ParentID)
		if err != nil {
			return commentLookupError(err)
		}
		if parent.PostID != postID {
			return errcode.ErrCommentNotFound
		}
		parentAuthor = parent.UserID
		if parent.RootID == 0 {
			rootID = parent.ID
		} else {
			rootID = parent.RootID
			replyUserID = parent.UserID
		}
	}
	comment := &model.PostComment{
		PostID:      postID,
		UserID:      userID,
		RootID:      rootID,
		ReplyUserID: replyUserID,
		Content:     content,
	}
	if err := s.repo.CreateComment(ctx, comment); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}

	// 通知（副作用）：回复→通知被回复评论作者；顶层评论→通知帖子作者
	link := "/blog/" + strconv.FormatInt(postID, 10)
	if req.ParentID > 0 {
		s.notify.Notify(ctx, parentAuthor, userID, "reply", "", content, link, "post", postID)
	} else {
		s.notify.Notify(ctx, post.UserID, userID, "comment", post.Title, content, link, "post", postID)
	}
	return nil
}

func (s *PostService) DeleteComment(ctx context.Context, commentID, userID int64, role int) error {
	comment, err := s.repo.GetComment(ctx, commentID)
	if err != nil {
		return commentLookupError(err)
	}
	if !canManagePost(comment.UserID, userID, role) {
		return errcode.ErrOperationDenied
	}
	if err := s.repo.DeleteComment(ctx, comment.ID, comment.PostID); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

// resolvePostProblem 校验/解析帖子关联题目：传入对外题号，返回内部主键（0=不关联）。
func (s *PostService) resolvePostProblem(ctx context.Context, category string, problemDisplayID string) (int64, error) {
	if problemDisplayID != "" {
		if category != postCategorySolution && category != postHelpSolution {
			return 0, nil // 非题解/求助不关联题目
		}
		pk, err := s.problemRepo.ResolveID(ctx, problemDisplayID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, errcode.ErrProblemNotFound
			}
			return 0, errcode.ErrDatabase.Wrap(err)
		}
		return pk, nil
	}
	// 未关联题目
	if category == postCategorySolution {
		return 0, errcode.ErrInvalidParams.WithMsg("题解必须关联题目")
	}
	return 0, nil
}

func (s *PostService) buildPostItems(ctx context.Context, posts []model.Post) ([]dto.PostItem, error) {
	userIDs := make([]int64, 0, len(posts))
	problemIDs := make([]int64, 0)
	for _, post := range posts {
		userIDs = append(userIDs, post.UserID)
		if post.ProblemID > 0 {
			problemIDs = append(problemIDs, post.ProblemID)
		}
	}

	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	userMap := make(map[int64]model.User, len(users))
	for _, user := range users {
		userMap[user.ID] = user
	}

	problems, err := s.problemRepo.FindByIDs(ctx, problemIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	problemMap := make(map[int64]model.Problem, len(problems))
	for _, problem := range problems {
		problemMap[problem.ID] = problem
	}

	items := make([]dto.PostItem, 0, len(posts))
	for _, post := range posts {
		item := dto.PostItem{
			ID:           post.ID,
			Category:     post.Category,
			ReviewStatus: post.ReviewStatus,
			RejectReason: post.RejectReason,
			// ProblemID 用对外题号，在下方按 problemMap 回填；避免泄露自增主键
			Title:        post.Title,
			Summary:      post.Summary,
			LikeCount:    post.LikeCount,
			ViewCount:    post.ViewCount,
			CommentCount: post.CommentCount,
			CreatedAt:    time.Unix(post.CreatedAt, 0).Format("2006-01-02 15:04:05"),
			UpdatedAt:    time.Unix(post.UpdatedAt, 0).Format("2006-01-02 15:04:05"),
		}
		if user, ok := userMap[post.UserID]; ok {
			item.User = dto.PostUser{ID: user.UID, Username: user.Username, Avatar: user.Avatar, Rating: user.Rating}
		}
		if problem, ok := problemMap[post.ProblemID]; ok {
			item.Problem = &dto.ProblemSimple{ID: problem.DisplayID, Name: problem.Name}
			item.ProblemID = problem.DisplayID
		}
		items = append(items, item)
	}
	return items, nil
}

// buildCommentTree 把扁平评论组装成单层楼中楼：顶层评论 + 各自的回复，回复带 ReplyTo。
func (s *PostService) buildCommentTree(ctx context.Context, comments []model.PostComment, likedSet map[int64]bool) ([]dto.PostCommentItem, error) {
	userIDs := make([]int64, 0, len(comments))
	for _, comment := range comments {
		userIDs = append(userIDs, comment.UserID)
	}
	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	userMap := make(map[int64]model.User, len(users))
	for _, user := range users {
		userMap[user.ID] = user
	}
	userOf := func(id int64) dto.PostUser {
		if u, ok := userMap[id]; ok {
			return dto.PostUser{ID: u.UID, Username: u.Username, Avatar: u.Avatar, Rating: u.Rating}
		}
		return dto.PostUser{} // 用户不存在：不暴露任何 id
	}

	// 一级评论(root_id=0)按顺序放好，二级评论按 root_id 挂到对应一级评论下
	roots := make([]dto.PostCommentItem, 0)
	rootIdx := make(map[int64]int) // 一级评论 id -> roots 下标
	for i := range comments {
		c := &comments[i]
		if c.RootID != 0 {
			continue
		}
		roots = append(roots, dto.PostCommentItem{
			ID:        c.ID,
			PostID:    c.PostID,
			Content:   c.Content,
			User:      userOf(c.UserID),
			LikeCount: c.LikeCount,
			IsLiked:   likedSet[c.ID],
			CreatedAt: c.CreatedAt,
			Replies:   []dto.PostCommentItem{},
		})
		rootIdx[c.ID] = len(roots) - 1
	}
	for i := range comments {
		c := &comments[i]
		if c.RootID == 0 {
			continue
		}
		idx, ok := rootIdx[c.RootID]
		if !ok {
			continue // 所属一级评论已被删，跳过孤儿回复
		}
		var replyTo *dto.PostUser
		if c.ReplyUserID != 0 {
			u := userOf(c.ReplyUserID)
			replyTo = &u
		}
		roots[idx].Replies = append(roots[idx].Replies, dto.PostCommentItem{
			ID:        c.ID,
			PostID:    c.PostID,
			Content:   c.Content,
			User:      userOf(c.UserID),
			ReplyTo:   replyTo,
			LikeCount: c.LikeCount,
			IsLiked:   likedSet[c.ID],
			CreatedAt: c.CreatedAt,
		})
	}
	return roots, nil
}

func normalizePostPage(req *dto.PostListReq) {
	if req.Page <= 0 {
		req.Page = defaultPostPage
	}
	if req.PageSize <= 0 || req.PageSize > maxPostPageSize {
		req.PageSize = defaultPostPageSize
	}
}

func canManagePost(ownerID, userID int64, role int) bool {
	return ownerID == userID || role == 1
}

const summaryMaxRunes = 120

// Markdown 清洗规则：作者不填摘要时，从正文自动截取。先剥掉常见 Markdown 标记，
// 否则摘要里会混入 #、**、代码块反引号、图片/链接语法等，很难看。
var (
	reCodeFence  = regexp.MustCompile("(?s)```.*?```") // 围栏代码块整段丢弃
	reImage      = regexp.MustCompile(`!\[[^\]]*\]\([^)]*\)`)
	reLink       = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`) // [文字](链接) -> 文字
	reInlineCode = regexp.MustCompile("`([^`]*)`")             // `code` -> code
	reHTMLTag    = regexp.MustCompile(`<[^>]+>`)
	reLinePrefix = regexp.MustCompile(`(?m)^\s{0,3}(#{1,6}\s*|>+\s*|[-*+]\s+|\d+\.\s+|[-*_]{3,}\s*)`) // 标题/引用/列表/分割线
	reEmphasis   = regexp.MustCompile(`[*_~]`)                                                        // 加粗/斜体/删除线标记
)

func makePostSummary(content string) string {
	text := content
	text = reCodeFence.ReplaceAllString(text, " ")
	text = reImage.ReplaceAllString(text, " ")
	text = reLink.ReplaceAllString(text, "$1")
	text = reInlineCode.ReplaceAllString(text, "$1")
	text = reHTMLTag.ReplaceAllString(text, " ")
	text = reLinePrefix.ReplaceAllString(text, "")
	text = reEmphasis.ReplaceAllString(text, "")
	// 折叠所有空白（含换行）为单个空格
	text = strings.Join(strings.Fields(text), " ")

	runes := []rune(text)
	if len(runes) > summaryMaxRunes {
		return strings.TrimSpace(string(runes[:summaryMaxRunes])) + "…"
	}
	return text
}
