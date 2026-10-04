package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/errcode"

	"gorm.io/gorm"
)

type TeamService struct {
	repo     *repository.TeamRepo
	userRepo *repository.UserRepo
}

func NewTeamService(repo *repository.TeamRepo, userRepo *repository.UserRepo) *TeamService {
	return &TeamService{repo: repo, userRepo: userRepo}
}

// ===== 查询 / 命令输入 =====

// TeamListParams 团队列表查询。Mine=1 只返回我加入的，Visibility 不传时返回全部权限类型。
type TeamListParams struct {
	Page       int
	PageSize   int
	Keyword    string
	Mine       int
	Visibility *int
}

// SaveTeamParams 新建/编辑团队。
type SaveTeamParams struct {
	Name        string
	CoverURL    string
	Description string
	Visibility  int
	InviteCode  string
}

// ===== 结果 =====

// TeamListItem 列表项。列表不返回简介，简介只在详情页展示。
type TeamListItem struct {
	ID          int64
	Name        string
	CoverURL    string
	Visibility  int
	MemberCount int
	Owner       string
	OwnerAvatar string
	MyRole      *int
	CreatedAt   time.Time
}

type TeamListResult struct {
	Total int64
	List  []TeamListItem
}

// TeamDetail 团队详情。
type TeamDetail struct {
	ID            int64
	Name          string
	CoverURL      string
	Description   string
	Visibility    int
	MemberCount   int
	HomeworkCount int
	Owner         string
	OwnerUID      int64
	MyRole        *int
	CanManage     bool
	InviteCode    string
	CreatedAt     time.Time
}

// TeamMemberItem 成员项。UID 是对外用户号，不暴露内部主键。
type TeamMemberItem struct {
	UID       int64
	Username  string
	StudentNo string
	RealName  string
	Gender    int
	Avatar    string
	Role      int
	JoinedAt  time.Time
}

type TeamMemberListResult struct {
	Total int
	List  []TeamMemberItem
}

// StudentImportResult 学生名单导入结果。
type StudentImportResult struct {
	Total          int
	CreatedUsers   int
	AddedMembers   int
	SkippedMembers int
}

// Access 取调用方在某团队中的权限视图。userID 为 0 表示未登录。
func (s *TeamService) Access(ctx context.Context, teamID, userID int64, isSiteAdmin bool) (TeamAccess, error) {
	access := TeamAccess{UserID: userID, IsSiteAdmin: isSiteAdmin}
	if userID == 0 {
		return access, nil
	}
	member, err := s.repo.GetMember(ctx, teamID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return access, nil
		}
		return access, errcode.ErrDatabase.Wrap(err)
	}
	access.IsMember = true
	access.Role = member.Role
	return access, nil
}

// List 团队分页列表。未登录也能看（列表本身不含团队内部内容）。
func (s *TeamService) List(ctx context.Context, params TeamListParams, userID int64) (*TeamListResult, error) {
	if params.Visibility != nil && *params.Visibility != model.TeamPublic && *params.Visibility != model.TeamInviteOnly {
		return nil, errcode.ErrInvalidParams.WithMsg("团队权限筛选值无效")
	}
	q := &repository.TeamQuery{
		Page:       params.Page,
		PageSize:   params.PageSize,
		Keyword:    params.Keyword,
		Visibility: params.Visibility,
	}
	if params.Mine == 1 {
		if userID == 0 {
			// 未登录时「我加入的」必然为空，直接返回，不必查库
			return &TeamListResult{Total: 0, List: []TeamListItem{}}, nil
		}
		q.MemberOf = userID
	}

	teams, total, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	resp := &TeamListResult{Total: total, List: make([]TeamListItem, 0, len(teams))}
	if len(teams) == 0 {
		return resp, nil
	}

	teamIDs := make([]int64, 0, len(teams))
	for _, t := range teams {
		teamIDs = append(teamIDs, t.ID)
	}

	// 成员数和我的角色都一次批量取回，避免逐个团队查询
	counts, err := s.repo.MemberCountsByTeamIDs(ctx, teamIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	countByTeam := make(map[int64]int, len(counts))
	for _, c := range counts {
		countByTeam[c.TeamID] = c.Count
	}
	roles, err := s.repo.MyRolesByTeamIDs(ctx, userID, teamIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	owners, err := s.ownerProfiles(ctx, teams)
	if err != nil {
		return nil, err
	}

	for _, t := range teams {
		owner := owners[t.OwnerID]
		item := TeamListItem{
			ID:          t.PublicID,
			Name:        t.Name,
			CoverURL:    t.CoverURL,
			Visibility:  t.Visibility,
			MemberCount: countByTeam[t.ID],
			Owner:       owner.Name,
			OwnerAvatar: owner.Avatar,
			CreatedAt:   t.CreatedAt,
		}
		if role, ok := roles[t.ID]; ok {
			item.MyRole = &role
		}
		resp.List = append(resp.List, item)
	}
	return resp, nil
}

// Detail 团队详情。名称/简介/成员数对任何人可见（含非公开团队的非成员），
// 这样别人知道团队存在、也知道它是干什么的；内部内容另由各自接口鉴权。
func (s *TeamService) Detail(ctx context.Context, id, userID int64, isSiteAdmin bool) (*TeamDetail, error) {
	team, err := s.getTeam(ctx, id)
	if err != nil {
		return nil, err
	}
	access, err := s.Access(ctx, id, userID, isSiteAdmin)
	if err != nil {
		return nil, err
	}

	counts, err := s.repo.MemberCountsByTeamIDs(ctx, []int64{id})
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	memberCount := 0
	if len(counts) > 0 {
		memberCount = counts[0].Count
	}
	hwCount, err := s.repo.HomeworkCount(ctx, id)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	owners, err := s.ownerProfiles(ctx, []model.Team{*team})
	if err != nil {
		return nil, err
	}

	resp := &TeamDetail{
		ID:            team.PublicID,
		Name:          team.Name,
		CoverURL:      team.CoverURL,
		Description:   team.Description,
		Visibility:    team.Visibility,
		MemberCount:   memberCount,
		HomeworkCount: int(hwCount),
		Owner:         owners[team.OwnerID].Name,
		CanManage:     access.CanManage(),
		CreatedAt:     team.CreatedAt,
	}
	if access.CanManage() {
		resp.InviteCode = team.InviteCode
	}
	if owner, err := s.userRepo.GetUserByID(ctx, team.OwnerID); err == nil {
		resp.OwnerUID = owner.UID
	}
	if access.IsMember {
		role := access.Role
		resp.MyRole = &role
	}
	return resp, nil
}

// Create 建团队，创建者自动成为所有者。
func (s *TeamService) Create(ctx context.Context, params SaveTeamParams, ownerID int64) (int64, error) {
	if strings.TrimSpace(params.Name) == "" {
		return 0, errcode.ErrInvalidParams.WithMsg("团队名称不能为空")
	}
	if err := validateTeamVisibility(params); err != nil {
		return 0, err
	}
	publicID, err := newUniquePublicID(ctx, s.repo.PublicIDExists)
	if err != nil {
		return 0, err
	}
	team := &model.Team{
		PublicID:    publicID,
		Name:        strings.TrimSpace(params.Name),
		CoverURL:    strings.TrimSpace(params.CoverURL),
		Description: params.Description,
		Visibility:  params.Visibility,
		InviteCode:  strings.TrimSpace(params.InviteCode),
		OwnerID:     ownerID,
	}
	if err := s.repo.CreateWithOwner(ctx, team); err != nil {
		return 0, errcode.ErrDatabase.Wrap(err)
	}
	return team.PublicID, nil
}

// Update 改团队信息，所有者和团队管理员可用。
func (s *TeamService) Update(ctx context.Context, id int64, params SaveTeamParams, userID int64, isSiteAdmin bool) error {
	if strings.TrimSpace(params.Name) == "" {
		return errcode.ErrInvalidParams.WithMsg("团队名称不能为空")
	}
	if err := validateTeamVisibility(params); err != nil {
		return err
	}
	if _, err := s.getTeam(ctx, id); err != nil {
		return err
	}
	access, err := s.Access(ctx, id, userID, isSiteAdmin)
	if err != nil {
		return err
	}
	if !access.CanManage() {
		return errcode.ErrPermissionDenied.WithMsg("无权修改该团队")
	}
	team := &model.Team{
		ID:          id,
		Name:        strings.TrimSpace(params.Name),
		CoverURL:    strings.TrimSpace(params.CoverURL),
		Description: params.Description,
		Visibility:  params.Visibility,
		InviteCode:  strings.TrimSpace(params.InviteCode),
	}
	if err := s.repo.Update(ctx, team); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

// Delete 解散团队，仅所有者可用。
func (s *TeamService) Delete(ctx context.Context, id, userID int64, isSiteAdmin bool) error {
	if _, err := s.getTeam(ctx, id); err != nil {
		return err
	}
	access, err := s.Access(ctx, id, userID, isSiteAdmin)
	if err != nil {
		return err
	}
	if !access.CanOwn() {
		return errcode.ErrPermissionDenied.WithMsg("只有团队所有者能解散团队")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

// Join 加入团队。公开团队直接进；非公开团队要邀请码对得上。
func (s *TeamService) Join(ctx context.Context, id, userID int64, code string) error {
	team, err := s.getTeam(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.repo.GetMember(ctx, id, userID); err == nil {
		return nil // 已在队里，幂等返回成功
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.ErrDatabase.Wrap(err)
	}

	if team.Visibility == model.TeamInviteOnly {
		if team.InviteCode == "" || team.InviteCode != strings.TrimSpace(code) {
			return errcode.ErrContestInviteCodeInvalid
		}
	}
	if err := s.repo.AddMember(ctx, id, userID); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

// Quit 退出团队。所有者不能退（要先转让或解散）。
func (s *TeamService) Quit(ctx context.Context, id, userID int64) error {
	if _, err := s.getTeam(ctx, id); err != nil {
		return err
	}
	// 退队只看自己的真实成员身份，站点超管身份不参与判断
	access, err := s.Access(ctx, id, userID, false)
	if err != nil {
		return err
	}
	if err := canQuitTeam(access); err != nil {
		return err
	}
	if err := s.repo.RemoveMember(ctx, id, userID); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

// ListMembers 成员列表，仅团队成员可见。
func (s *TeamService) ListMembers(ctx context.Context, id, userID int64, isSiteAdmin bool) (*TeamMemberListResult, error) {
	if _, err := s.getTeam(ctx, id); err != nil {
		return nil, err
	}
	access, err := s.Access(ctx, id, userID, isSiteAdmin)
	if err != nil {
		return nil, err
	}
	if !access.CanView() {
		return nil, errcode.ErrPermissionDenied.WithMsg("仅团队成员可查看")
	}

	members, err := s.repo.ListMembers(ctx, id)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	resp := &TeamMemberListResult{Total: len(members), List: make([]TeamMemberItem, 0, len(members))}
	if len(members) == 0 {
		return resp, nil
	}

	userIDs := make([]int64, 0, len(members))
	for _, m := range members {
		userIDs = append(userIDs, m.UserID)
	}
	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	userByID := make(map[int64]model.User, len(users))
	for _, u := range users {
		userByID[u.ID] = u
	}

	for _, m := range members {
		u := userByID[m.UserID]
		resp.List = append(resp.List, TeamMemberItem{
			UID:       u.UID, // 对外用户号，不暴露内部主键
			Username:  u.Username,
			StudentNo: studentNoValue(u.StudentNo),
			RealName:  u.RealName,
			Gender:    u.Gender,
			Avatar:    u.Avatar,
			Role:      m.Role,
			JoinedAt:  m.JoinedAt,
		})
	}
	return resp, nil
}

// SetMemberRole 设置/取消团队管理员，仅所有者可用。uid 为对外用户号。
func (s *TeamService) SetMemberRole(ctx context.Context, id, uid, actorID int64, isSiteAdmin bool, role int) error {
	if _, err := s.getTeam(ctx, id); err != nil {
		return err
	}
	access, err := s.Access(ctx, id, actorID, isSiteAdmin)
	if err != nil {
		return err
	}
	if !access.CanOwn() {
		return errcode.ErrPermissionDenied.WithMsg("只有团队所有者能设置管理员")
	}
	targetID, target, err := s.resolveMember(ctx, id, uid)
	if err != nil {
		return err
	}
	if target.Role == model.TeamRoleOwner {
		return errcode.ErrOperationDenied.WithMsg("不能修改团队所有者的角色")
	}
	if err := s.repo.SetMemberRole(ctx, id, targetID, role); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

// RemoveMember 移出成员。uid 为对外用户号。
func (s *TeamService) RemoveMember(ctx context.Context, id, uid, actorID int64, isSiteAdmin bool) error {
	if _, err := s.getTeam(ctx, id); err != nil {
		return err
	}
	access, err := s.Access(ctx, id, actorID, isSiteAdmin)
	if err != nil {
		return err
	}
	targetID, target, err := s.resolveMember(ctx, id, uid)
	if err != nil {
		return err
	}
	if err := canRemoveMember(access, target.Role); err != nil {
		return err
	}
	if err := s.repo.RemoveMember(ctx, id, targetID); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

// ===== 内部工具 =====

// ResolveID 把路由中的公开团队编号解析为数据库内部主键。
func (s *TeamService) ResolveID(ctx context.Context, publicID int64) (int64, error) {
	id, err := s.repo.ResolveIDByPublicID(ctx, publicID)
	if err == nil {
		return id, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errcode.ErrTeamNotFound
	}
	return 0, errcode.ErrDatabase.Wrap(err)
}

func (s *TeamService) getTeam(ctx context.Context, id int64) (*model.Team, error) {
	team, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrTeamNotFound
		}
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	return team, nil
}

// resolveMember 把对外用户号解析成内部主键，并确认其确实是该团队成员。
func (s *TeamService) resolveMember(ctx context.Context, teamID, uid int64) (int64, *model.TeamMember, error) {
	targetID, err := s.userRepo.ResolveIDByUID(ctx, uid)
	if err != nil {
		return 0, nil, errcode.ErrUserNotFound
	}
	member, err := s.repo.GetMember(ctx, teamID, targetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil, errcode.ErrOperationDenied.WithMsg("该用户不是团队成员")
		}
		return 0, nil, errcode.ErrDatabase.Wrap(err)
	}
	return targetID, member, nil
}

type teamOwnerProfile struct {
	Name   string
	Avatar string
}

func (s *TeamService) ownerProfiles(ctx context.Context, teams []model.Team) (map[int64]teamOwnerProfile, error) {
	ids := make([]int64, 0, len(teams))
	seen := make(map[int64]struct{}, len(teams))
	for _, t := range teams {
		if t.OwnerID == 0 {
			continue
		}
		if _, ok := seen[t.OwnerID]; ok {
			continue
		}
		seen[t.OwnerID] = struct{}{}
		ids = append(ids, t.OwnerID)
	}
	if len(ids) == 0 {
		return map[int64]teamOwnerProfile{}, nil
	}
	users, err := s.userRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	profiles := make(map[int64]teamOwnerProfile, len(users))
	for _, u := range users {
		profiles[u.ID] = teamOwnerProfile{Name: u.Username, Avatar: u.Avatar}
	}
	return profiles, nil
}

func validateTeamVisibility(params SaveTeamParams) error {
	if params.Visibility == model.TeamInviteOnly && strings.TrimSpace(params.InviteCode) == "" {
		return errcode.ErrInvalidParams.WithMsg("非公开团队必须设置邀请码")
	}
	return nil
}
