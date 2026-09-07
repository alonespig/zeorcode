package service

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"zoj/internal/common/errcode"
	"zoj/internal/dto"
	"zoj/internal/infra/cache"
	"zoj/internal/middleware"
	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/util"

	"gorm.io/gorm"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{2,20}$`)

func registrationRealName(realName, username string) string {
	realName = strings.TrimSpace(realName)
	if realName == "" {
		return username
	}
	return realName
}

func validatePassword(password string) error {
	if utf8.RuneCountInString(password) < 6 {
		return errcode.ErrInvalidParams.WithMsg("密码至少 6 位")
	}
	// bcrypt 只接受不超过 72 字节的明文；中文等字符可能占多个字节。
	if len(password) > 72 {
		return errcode.ErrInvalidParams.WithMsg("密码不能超过 72 个字节")
	}
	return nil
}

// randUID 生成一个 8 位随机用户号（10000000 ~ 99999999）。
func randUID() int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(90000000))
	if err != nil {
		return 10000000 + time.Now().UnixNano()%90000000 // 兜底，基本不会走到
	}
	return 10000000 + n.Int64()
}

// ResolveUID 把对外用户号解析成内部主键（handler 用它把 URL 里的 uid 换成主键）。
func (u *UserService) ResolveUID(ctx context.Context, uid int64) (int64, error) {
	id, err := u.repo.ResolveIDByUID(ctx, uid)
	if err != nil {
		return 0, errcode.ErrUserNotFound
	}
	return id, nil
}

// genUniqueUID 生成不与现有用户冲突的 uid（碰撞极少，重试几次）。
func (u *UserService) genUniqueUID(ctx context.Context) (int64, error) {
	return genUniqueUID(ctx, u.repo)
}

// genUniqueUID 供注册、后台批量创建和团队名单导入共用。
func genUniqueUID(ctx context.Context, repo *repository.UserRepo) (int64, error) {
	for i := 0; i < 10; i++ {
		uid := randUID()
		exists, err := repo.ExistsByUID(ctx, uid)
		if err != nil {
			return 0, errcode.ErrDatabase.Wrap(err)
		}
		if !exists {
			return uid, nil
		}
	}
	return 0, errcode.ErrInternal.WithMsg("生成用户号失败，请重试")
}

// defaultAvatar 用户未设置头像时的占位图
const defaultAvatar = "https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png"

// userRankTTL 用户排行榜缓存有效期（兜底，主要靠新 AC 时换代失效）
const userRankTTL = 30 * time.Second

func avatarOr(a string) string {
	if a == "" {
		return defaultAvatar
	}
	return a
}

func loginResponse(user *model.User) *dto.LoginResp {
	return &dto.LoginResp{
		User: dto.UserInfo{
			ID:       user.UID,
			Username: user.Username,
			Role:     user.Role,
		},
		Avatar: avatarOr(user.Avatar),
	}
}

type UserService struct {
	repo        *repository.UserRepo
	subRepo     *repository.SubmissionRepo
	problemRepo *repository.ProblemRepo
	cache       *cache.Cache
	auth        *middleware.Auth
	verify      *VerifyService
}

func NewUserService(repo *repository.UserRepo,
	subRepo *repository.SubmissionRepo,
	problemRepo *repository.ProblemRepo,
	cache *cache.Cache,
	auth *middleware.Auth,
	verify *VerifyService) *UserService {
	return &UserService{
		repo:        repo,
		subRepo:     subRepo,
		problemRepo: problemRepo,
		cache:       cache,
		auth:        auth,
		verify:      verify,
	}
}

func (u *UserService) CreateUser(ctx context.Context, req *dto.CreateUserReq) error {
	username := strings.TrimSpace(req.Username)
	studentNo := strings.TrimSpace(req.StudentNo)
	realName := registrationRealName(req.RealName, username)
	email := normalizeEmail(req.Email)
	avatar := strings.TrimSpace(req.Avatar)
	if !usernamePattern.MatchString(username) {
		return errcode.ErrInvalidParams.WithMsg("用户名只能包含字母、数字和下划线，长度为 2-20 位")
	}
	if email == "" {
		return errcode.ErrInvalidParams.WithMsg("邮箱不能为空")
	}
	if err := validatePassword(req.Password); err != nil {
		return err
	}

	// 查重：用户名、学号和邮箱分别返回可识别的错误。
	if exists, err := u.repo.ExistsByUsername(ctx, username); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	} else if exists {
		return errcode.ErrUserExists.WithMsg("用户名已被注册")
	}
	if studentNo != "" {
		if exists, err := u.repo.ExistsByStudentNo(ctx, studentNo); err != nil {
			return errcode.ErrDatabase.Wrap(err)
		} else if exists {
			return errcode.ErrUserExists.WithMsg("学号已存在")
		}
	}
	if exists, err := u.repo.ExistsByEmail(ctx, email); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	} else if exists {
		return errcode.ErrEmailExists
	}

	// 唯一性检查通过后再校验验证码（校验成功即消费掉，避免因用户名占用等原因白白消耗）
	if err := u.verify.Verify(ctx, SceneRegister, email, req.Code); err != nil {
		return err
	}

	hash, err := util.HashPassword(req.Password)
	if err != nil {
		return err
	}
	uid, err := u.genUniqueUID(ctx)
	if err != nil {
		return err
	}
	return u.repo.CreateUser(ctx, &model.User{
		UID:       uid,
		Username:  username,
		StudentNo: studentNoPtr(studentNo),
		RealName:  realName,
		Password:  hash,
		Email:     emailPtr(email),
		Avatar:    avatar,
	})
}

// ResetPassword 通过邮箱验证码重置密码。
func (u *UserService) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	email = normalizeEmail(email)
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	if err := u.verify.Verify(ctx, SceneReset, email, code); err != nil {
		return err
	}
	user, err := u.repo.GetUserByLogin(ctx, email)
	if err != nil {
		return errcode.ErrUserNotFound.WithMsg("该邮箱未注册")
	}
	hash, err := util.HashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := u.repo.UpdatePassword(ctx, user.ID, hash); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	// 找回密码后撤销已有会话，避免旧登录态继续访问账号。
	_ = u.auth.Revoke(ctx, user.ID)
	return nil
}

// ChangePassword 校验当前密码后修改密码，并撤销当前会话。
func (u *UserService) ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	if currentPassword == newPassword {
		return errcode.ErrInvalidParams.WithMsg("新密码不能与当前密码相同")
	}
	user, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		return errcode.ErrUserNotFound
	}
	if !util.CheckPassword(currentPassword, user.Password) {
		return errcode.ErrWrongPassword.WithMsg("当前密码错误")
	}
	hash, err := util.HashPassword(newPassword)
	if err != nil {
		return errcode.ErrInternal.Wrap(err)
	}
	if err := u.repo.UpdatePassword(ctx, userID, hash); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	_ = u.auth.Revoke(ctx, userID)
	return nil
}

// BindEmail 换绑/绑定邮箱（需登录）：校验发往新邮箱的验证码后更新。
func (u *UserService) BindEmail(ctx context.Context, userID int64, email, code string) error {
	email = normalizeEmail(email)
	if err := u.verify.Verify(ctx, SceneBind, email, code); err != nil {
		return err
	}
	// 二次确认新邮箱未被别人占用（验证码有效期内可能被抢注）
	if exists, err := u.repo.ExistsByEmail(ctx, email); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	} else if exists {
		return errcode.ErrEmailExists.WithMsg("该邮箱已被占用")
	}
	user, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		return errcode.ErrUserNotFound
	}
	user.Email = emailPtr(email)
	if err := u.repo.UpdateUser(ctx, user); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

func (u *UserService) UserRankList(ctx context.Context, page, pageSize int, username *string) (*dto.UserRank, error) {

	filter := repository.UserSubmitCountFilter{
		Page:     page,
		PageSize: pageSize,
	}

	if username != nil {
		filter.QueryName = true
		users, err := u.repo.SearchByUsername(ctx, *username)
		if err != nil {
			return nil, errcode.ErrDatabase.Wrap(err)
		}
		for _, user := range users {
			filter.UserIDs = append(filter.UserIDs, int(user.ID))
		}
	}
	rank, total, err := u.problemRepo.GetUserSubmitCount(ctx, &filter)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}

	resp := dto.UserRank{
		Total:    total,
		UserRank: make([]dto.UserInfo, len(rank)),
	}
	// 批量取用户信息，避免每个排名一条查询（N+1）
	userIDs := make([]int64, 0, len(rank))
	for _, item := range rank {
		userIDs = append(userIDs, item.UserID)
	}
	users, err := u.repo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	userMap := make(map[int64]model.User, len(users))
	for _, usr := range users {
		userMap[usr.ID] = usr
	}
	for i, item := range rank {
		user := userMap[item.UserID]
		resp.UserRank[i] = dto.UserInfo{
			ID:          user.UID,
			Index:       (page-1)*pageSize + i + 1,
			Username:    user.Username,
			PassCount:   item.ACCount,
			SubmitCount: item.SubmitCount,
			Signature:   user.Signature,
			Avatar:      avatarOr(user.Avatar),
			Rating:      user.Rating,
			MaxRating:   user.MaxRating,
		}
	}

	return &resp, nil
}

// RatingRankList 按 rating 降序的全站排名。
func (u *UserService) RatingRankList(ctx context.Context, page, pageSize int) (*dto.UserRank, error) {
	users, total, err := u.repo.RatingRankList(ctx, page, pageSize)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	resp := &dto.UserRank{Total: total, UserRank: make([]dto.UserInfo, len(users))}
	for i, usr := range users {
		resp.UserRank[i] = dto.UserInfo{
			ID:        usr.UID,
			Index:     (page-1)*pageSize + i + 1,
			Username:  usr.Username,
			Avatar:    avatarOr(usr.Avatar),
			Signature: usr.Signature,
			Rating:    usr.Rating,
			MaxRating: usr.MaxRating,
		}
	}
	return resp, nil
}

// GetRatingHistory 某用户当前 rating + 历次 rating 变化（个人页折线图用）。
func (u *UserService) GetRatingHistory(ctx context.Context, userID int64) (*dto.RatingHistoryResp, error) {
	user, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errcode.ErrUserNotFound
	}
	rows, err := u.repo.GetRatingHistory(ctx, userID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	resp := &dto.RatingHistoryResp{
		Rating:    user.Rating,
		MaxRating: user.MaxRating,
		History:   make([]dto.RatingHistoryItem, len(rows)),
	}
	for i, r := range rows {
		resp.History[i] = dto.RatingHistoryItem{
			ContestID:   r.ContestID,
			ContestName: r.ContestName,
			Rank:        r.Rank,
			OldRating:   r.OldRating,
			NewRating:   r.NewRating,
			Delta:       r.Delta,
			Time:        r.CreatedAt.UnixMilli(),
		}
	}
	return resp, nil
}

// GetContestHistory 个人页「比赛记录」：用户报名过的比赛 + 每场结算状态（settled 带 rating 变化）。
func (u *UserService) GetContestHistory(ctx context.Context, userID int64) (*dto.ContestHistoryResp, error) {
	rows, err := u.repo.GetContestHistory(ctx, userID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	now := time.Now()
	resp := &dto.ContestHistoryResp{List: make([]dto.ContestHistoryItem, 0, len(rows))}
	for _, r := range rows {
		item := dto.ContestHistoryItem{
			ContestID:   r.ContestID,
			ContestName: r.ContestName,
			Type:        r.Type,
			Time:        r.StartTime.Unix(),
		}
		switch {
		case r.NewRating != nil: // 有 rating 变化即已结算
			item.Status = "settled"
			item.Total = r.Total
			if r.Rank != nil {
				item.Rank = *r.Rank
			}
			if r.OldRating != nil {
				item.OldRating = *r.OldRating
			}
			item.NewRating = *r.NewRating
			if r.Delta != nil {
				item.Delta = *r.Delta
			}
		case now.Before(r.StartTime):
			item.Status = "upcoming"
		case now.Before(r.EndTime):
			item.Status = "running"
		case r.Rated: // 已结束 + 计入 rating 但还没结算
			item.Status = "calculating"
		default: // 已结束 + 不计分
			item.Status = "unrated"
		}
		resp.List = append(resp.List, item)
	}
	return resp, nil
}

func (u *UserService) UserInfo(ctx context.Context, id int64) (*dto.UserInfoResp, error) {
	user, err := u.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := dto.UserInfoResp{
		ID:          user.UID,
		Name:        user.Username,
		Avatar:      avatarOr(user.Avatar),
		Email:       emailValue(user.Email),
		Gender:      user.Gender,
		Signature:   &user.Signature,
		School:      user.School,
		CreatedAt:   user.CreatedAt.Unix(),
		SolveItem:   make([]dto.ProblemSimple, 0, 10),
		UnsolveItem: make([]dto.ProblemSimple, 0, 10),
	}

	acIds, unacIds, err := u.problemRepo.GetProblemIDByUserID(ctx, id)
	if err != nil {
		return nil, err
	}
	// 一次性取出所有相关题目，避免每道题一条查询（N+1）
	allIds := make([]int64, 0, len(acIds)+len(unacIds))
	allIds = append(allIds, acIds...)
	allIds = append(allIds, unacIds...)
	problems, err := u.problemRepo.FindByIDs(ctx, allIds)
	if err != nil {
		return nil, err
	}
	nameOf := make(map[int64]string, len(problems))
	displayOf := make(map[int64]string, len(problems)) // 主键 → 对外题号
	for _, p := range problems {
		nameOf[p.ID] = p.Name
		displayOf[p.ID] = p.DisplayID
	}
	for _, pid := range acIds {
		resp.SolveItem = append(resp.SolveItem, dto.ProblemSimple{
			ID:   displayOf[pid], // 对外题号，链接跳 /problem/{题号}
			Name: nameOf[pid],
		})
	}
	for _, pid := range unacIds {
		resp.UnsolveItem = append(resp.UnsolveItem, dto.ProblemSimple{
			ID:   displayOf[pid],
			Name: nameOf[pid],
		})
	}

	sort.Slice(resp.SolveItem, func(i, j int) bool {
		return resp.SolveItem[i].ID < resp.SolveItem[j].ID
	})
	return &resp, nil
}

func (u *UserService) Profile(ctx context.Context, id int64) (*dto.UserProfile, error) {
	user, err := u.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	acIds, unacIds, err := u.problemRepo.GetProblemIDByUserID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := dto.UserProfile{
		User: dto.UserDetail{
			ID:        user.UID,
			Name:      user.Username,
			Gender:    user.Gender,
			Signature: &user.Signature,
			School:    user.School,
			Avatar:    avatarOr(user.Avatar),
			CreatedAt: user.CreatedAt.Format("2006-6-6"),
		},
		SolveItem:   make([]dto.ProblemSimple, 0, len(acIds)),
		UnsolveItem: make([]dto.ProblemSimple, 0, len(unacIds)),
	}

	var (
		acProblemList   []model.Problem
		unacProblemList []model.Problem
	)

	acProblemList, err = u.problemRepo.FindByIDs(ctx, acIds)
	if err != nil {
		return nil, err
	}

	unacProblemList, err = u.problemRepo.FindByIDs(ctx, unacIds)
	if err != nil {
		return nil, err
	}

	for _, p := range acProblemList {
		resp.SolveItem = append(resp.SolveItem, dto.ProblemSimple{
			ID:   p.DisplayID, // 对外题号
			Name: p.Name,
		})
	}

	for _, p := range unacProblemList {
		resp.UnsolveItem = append(resp.UnsolveItem, dto.ProblemSimple{
			ID:   p.DisplayID,
			Name: p.Name,
		})
	}

	sort.Slice(resp.SolveItem, func(i, j int) bool {
		return resp.SolveItem[i].ID < resp.SolveItem[j].ID
	})
	return &resp, nil
}

// UpdateProfile 更新当前用户可编辑资料（签名、学校、性别、头像）；邮箱换绑走独立验证流程。
func (u *UserService) UpdateProfile(ctx context.Context, userID int64, req *dto.UpdateProfileReq) error {
	user, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		return errcode.ErrUserNotFound
	}
	user.Signature = req.Signature
	user.School = req.School
	user.Gender = req.Gender
	user.Avatar = req.Avatar
	// 邮箱不在这里改：换绑走 BindEmail（需邮箱验证码），避免绕过验证
	if err := u.repo.UpdateUser(ctx, user); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

// Login 支持用户名或邮箱登录（name 传任意一个）
func (u *UserService) Login(ctx context.Context, name, password string) (*dto.LoginResp, string, error) {
	user, err := u.repo.GetUserByLogin(ctx, name)
	if err != nil {
		return nil, "", errcode.ErrUserNotFound
	}
	if !util.CheckPassword(password, user.Password) {
		return nil, "", errcode.ErrWrongPassword
	}
	if user.Status == model.UserStatusBanned {
		return nil, "", errcode.ErrUserDisabled.WithMsg("账号已被封禁")
	}
	token, err := u.auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, "", errcode.ErrInternal.Wrap(err)
	}
	return loginResponse(user), token, nil
}

// Session 根据已通过 JWTAuthOptional 验证的内部用户主键返回当前会话用户。
// 前端用它在应用启动时把本地展示状态与 HttpOnly Cookie 同步。
func (u *UserService) Session(ctx context.Context, userID int64) (*dto.LoginResp, error) {
	user, err := u.repo.GetUserByID(ctx, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrUserNotFound
	}
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	if user.Status == model.UserStatusBanned {
		return nil, errcode.ErrUserDisabled.WithMsg("账号已被封禁")
	}
	return loginResponse(user), nil
}

// ListAllUsers 管理员视角的用户分页列表
func (u *UserService) ListAllUsers(ctx context.Context, page, pageSize int) (*dto.AdminUserListResp, error) {
	users, total, err := u.repo.ListUsers(ctx, page, pageSize)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	resp := &dto.AdminUserListResp{
		Total: total,
		List:  make([]dto.AdminUserItem, 0, len(users)),
	}
	for _, user := range users {
		resp.List = append(resp.List, dto.AdminUserItem{
			ID:        user.UID, // 对外用户号
			Username:  user.Username,
			StudentNo: studentNoValue(user.StudentNo),
			RealName:  user.RealName,
			Email:     emailValue(user.Email),
			Role:      user.Role,
			Status:    user.Status,
			Signature: user.Signature,
			CreatedAt: user.CreatedAt.Unix(),
		})
	}
	return resp, nil
}

// SetUserStatus 管理员封禁/解封用户（uid 为对外用户号）。
// 防呆：不能封禁管理员（含自己）。封禁时撤销其登录态，立即踢下线。
func (u *UserService) SetUserStatus(ctx context.Context, uid int64, status int) error {
	id, err := u.repo.ResolveIDByUID(ctx, uid)
	if err != nil {
		return errcode.ErrUserNotFound
	}
	user, err := u.repo.GetUserByID(ctx, id)
	if err != nil {
		return errcode.ErrUserNotFound
	}
	if user.Role == 1 {
		return errcode.ErrOperationDenied.WithMsg("不能封禁管理员")
	}
	user.Status = status
	if err := u.repo.UpdateUser(ctx, user); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	// 封禁：撤销其 session，现有 token 下次请求即失效（立即踢下线）
	if status == model.UserStatusBanned {
		_ = u.auth.Revoke(ctx, id)
	}
	return nil
}

// changeRolePolicy 修改角色的前置校验。抽成纯函数便于单测（同 contestProblemAccessPolicy）。
// actorID / targetID 均为内部主键。
func changeRolePolicy(actorID, targetID int64, newRole int) error {
	if actorID == targetID {
		return errcode.ErrOperationDenied.WithMsg("不能修改自己的角色")
	}
	// 兜底：DTO 的 binding:"oneof=0 1" 已挡过一层，这里防止其他调用方绕过
	if newRole != model.RoleNormal && newRole != model.RoleAdmin {
		return errcode.ErrInvalidParams.WithMsg("非法角色")
	}
	return nil
}

// SetUserRole 管理员修改用户角色（actorID 为操作者内部主键，uid 为目标用户的对外用户号）。
// 防呆：不能修改自己的角色，避免误操作后失去后台入口；管理员之间平级，可互相升降。
// 幂等：目标角色与当前一致时直接返回成功，不写库也不撤销会话——
// 对一个空操作把人踢下线是意外副作用。
// 角色变更后必须撤销会话：AdminRequired 只认 JWT 里签发时的 role 声明、不查库，
// 不撤销的话被降级者在 token 过期前仍能访问管理员接口；
// 升级同理，重新登录才能拿到带 role=1 的新 token。
//
// 已知限制（本次不修）：
//  1. 两个管理员并发互降时，各自都没碰"自己那一行"，校验都会通过，
//     若系统只有这两个管理员则可能降到零管理员。彻底修需要事务化的
//     管理员计数加锁校验，成本远高于收益；SetUserStatus 的封禁守卫存在同类竞态。
//  2. 可以把已封禁用户提升为管理员（封禁守卫只在封禁时查角色，不在改角色时查状态），
//     属于封禁功能本身的既有不对称。
func (u *UserService) SetUserRole(ctx context.Context, actorID, uid int64, role int) error {
	id, err := u.repo.ResolveIDByUID(ctx, uid)
	if err != nil {
		return errcode.ErrUserNotFound
	}
	// 策略校验放在读用户之前：只依赖 id/role，失败时省一次查询
	if err := changeRolePolicy(actorID, id, role); err != nil {
		return err
	}
	user, err := u.repo.GetUserByID(ctx, id)
	if err != nil {
		return errcode.ErrUserNotFound
	}
	if user.Role == role {
		return nil
	}
	user.Role = role
	if err := u.repo.UpdateUser(ctx, user); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	_ = u.auth.Revoke(ctx, id)
	return nil
}

// BatchCreateUsers 批量创建用户：
//   - 入参列表内用户名、学号、非空邮箱去重
//   - 与库内任一唯一身份冲突时全部拒绝（一次性返回失败）
//   - 成功则原子插入
func (u *UserService) BatchCreateUsers(ctx context.Context, req *dto.BatchCreateUsersReq) (*dto.BatchCreateUsersResp, error) {
	names := make([]string, 0, len(req.Users))
	emails := make([]string, 0, len(req.Users))
	studentNos := make([]string, 0, len(req.Users))
	seenNames := make(map[string]bool, len(req.Users))
	seenEmails := make(map[string]bool, len(req.Users))
	seenStudentNos := make(map[string]bool, len(req.Users))
	for i := range req.Users {
		item := &req.Users[i]
		item.Username = strings.TrimSpace(item.Username)
		item.StudentNo = strings.TrimSpace(item.StudentNo)
		item.RealName = strings.TrimSpace(item.RealName)
		item.Email = normalizeEmail(item.Email)
		if !usernamePattern.MatchString(item.Username) {
			return nil, errcode.ErrInvalidParams.WithMsg("用户名只能包含字母、数字和下划线，长度为 2-20 位: " + item.Username)
		}
		if item.RealName == "" {
			return nil, errcode.ErrInvalidParams.WithMsg("姓名不能为空")
		}
		if item.Role == model.RoleNormal && item.StudentNo == "" {
			return nil, errcode.ErrInvalidParams.WithMsg("普通用户必须填写学号: " + item.Username)
		}
		if seenNames[item.Username] {
			return nil, errcode.ErrInvalidParams.WithMsg("用户名重复: " + item.Username)
		}
		if item.Email != "" && seenEmails[item.Email] {
			return nil, errcode.ErrInvalidParams.WithMsg("邮箱重复: " + item.Email)
		}
		if item.StudentNo != "" && seenStudentNos[item.StudentNo] {
			return nil, errcode.ErrInvalidParams.WithMsg("学号重复: " + item.StudentNo)
		}
		seenNames[item.Username] = true
		names = append(names, item.Username)
		if item.Email != "" {
			seenEmails[item.Email] = true
			emails = append(emails, item.Email)
		}
		if item.StudentNo != "" {
			seenStudentNos[item.StudentNo] = true
			studentNos = append(studentNos, item.StudentNo)
		}
	}

	existing, err := u.repo.CountByUsernames(ctx, names)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	if len(existing) > 0 {
		return nil, errcode.ErrUserExists.WithMsg("用户名已存在: " + existing[0])
	}
	existing, err = u.repo.CountByEmails(ctx, emails)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	if len(existing) > 0 {
		return nil, errcode.ErrEmailExists.WithMsg("邮箱已存在: " + existing[0])
	}
	existing, err = u.repo.CountByStudentNos(ctx, studentNos)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	if len(existing) > 0 {
		return nil, errcode.ErrUserExists.WithMsg("学号已存在: " + existing[0])
	}

	users := make([]*model.User, 0, len(req.Users))
	for _, item := range req.Users {
		hash, err := util.HashPassword(item.Password)
		if err != nil {
			return nil, errcode.ErrInternal.Wrap(err)
		}
		uid, err := u.genUniqueUID(ctx)
		if err != nil {
			return nil, err
		}
		users = append(users, &model.User{
			UID:       uid,
			Username:  item.Username,
			StudentNo: studentNoPtr(item.StudentNo),
			RealName:  item.RealName,
			Email:     emailPtr(item.Email),
			Password:  hash,
			Role:      item.Role,
		})
	}
	if err := u.repo.CreateUsersBatch(ctx, users); err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	return &dto.BatchCreateUsersResp{Created: len(users)}, nil
}

func studentNoPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func emailPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func emailValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func studentNoValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
