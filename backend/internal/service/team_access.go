package service

import (
	"zoj/internal/common/errcode"
	"zoj/internal/model"
)

// TeamAccess 调用方在某个团队中的权限视图。
// 抽成纯类型 + 纯函数，便于在不碰数据库的前提下表驱动单测
// （同 contestProblemAccessPolicy 的做法）。
type TeamAccess struct {
	UserID      int64
	IsMember    bool
	Role        int  // model.TeamRole*，非成员时无意义
	IsSiteAdmin bool // 站点超管（users.role==1），拥有所有团队的全部权限
}

// CanView 能否查看团队内部内容（作业、成员、文件）。
func (a TeamAccess) CanView() bool {
	return a.IsMember || a.IsSiteAdmin
}

// CanManage 能否改团队信息、管成员、布置作业、查看全部成员的提交。
func (a TeamAccess) CanManage() bool {
	if a.IsSiteAdmin {
		return true
	}
	return a.IsMember && a.Role >= model.TeamRoleAdmin
}

// CanOwn 所有者专属：解散团队、设置/取消管理员。
func (a TeamAccess) CanOwn() bool {
	if a.IsSiteAdmin {
		return true
	}
	return a.IsMember && a.Role == model.TeamRoleOwner
}

// canRemoveMember 判断能否把 target 移出团队。
// 所有者不可被移出（要先转让或解散）；管理员之间不能互相移出，防止权限升级
// 后互踢，只有所有者能动管理员。
func canRemoveMember(actor TeamAccess, targetRole int) error {
	if !actor.CanManage() {
		return errcode.ErrPermissionDenied.WithMsg("无权管理团队成员")
	}
	if targetRole == model.TeamRoleOwner {
		return errcode.ErrOperationDenied.WithMsg("不能移出团队所有者")
	}
	if targetRole == model.TeamRoleAdmin && !actor.CanOwn() {
		return errcode.ErrOperationDenied.WithMsg("只有团队所有者能移出管理员")
	}
	return nil
}

// canQuitTeam 所有者不能直接退队，否则团队会没人能管理。
func canQuitTeam(a TeamAccess) error {
	if !a.IsMember {
		return errcode.ErrOperationDenied.WithMsg("不在该团队中")
	}
	if a.Role == model.TeamRoleOwner {
		return errcode.ErrOperationDenied.WithMsg("团队所有者不能退出团队，请先转让所有权或解散团队")
	}
	return nil
}

// canEditHomework 方案 B：只有作业的创建者本人能修改，团队管理员也不行。
// 站点超管例外（便于处理违规内容）。
func canEditHomework(a TeamAccess, hw *model.Homework) error {
	if a.IsSiteAdmin || a.UserID == hw.CreatedBy {
		return nil
	}
	return errcode.ErrOperationDenied.WithMsg("只有作业的布置者本人能修改该作业")
}

// canDeleteHomework 删除比修改多给团队所有者兜底：
// 否则布置人离队后，作业就成了没人能清理的僵尸数据。
func canDeleteHomework(a TeamAccess, hw *model.Homework) error {
	if a.IsSiteAdmin || a.UserID == hw.CreatedBy || (a.IsMember && a.Role == model.TeamRoleOwner) {
		return nil
	}
	return errcode.ErrOperationDenied.WithMsg("只有作业的布置者或团队所有者能删除该作业")
}
