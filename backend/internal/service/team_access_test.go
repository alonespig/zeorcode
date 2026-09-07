package service

import (
	"testing"

	"zoj/internal/common/errcode"
	"zoj/internal/model"
)

func member(role int) TeamAccess {
	return TeamAccess{UserID: 10, IsMember: true, Role: role}
}

func TestTeamAccessCapabilities(t *testing.T) {
	tests := []struct {
		name                          string
		access                        TeamAccess
		wantView, wantManage, wantOwn bool
	}{
		{"非成员什么都不能", TeamAccess{UserID: 10}, false, false, false},
		{"普通成员只能看", member(model.TeamRoleMember), true, false, false},
		{"团队管理员能管但不能解散", member(model.TeamRoleAdmin), true, true, false},
		{"所有者全能", member(model.TeamRoleOwner), true, true, true},
		{"站点超管即使不是成员也全能", TeamAccess{UserID: 99, IsSiteAdmin: true}, true, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.access.CanView(); got != tt.wantView {
				t.Errorf("CanView() = %v, want %v", got, tt.wantView)
			}
			if got := tt.access.CanManage(); got != tt.wantManage {
				t.Errorf("CanManage() = %v, want %v", got, tt.wantManage)
			}
			if got := tt.access.CanOwn(); got != tt.wantOwn {
				t.Errorf("CanOwn() = %v, want %v", got, tt.wantOwn)
			}
		})
	}
}

func TestCanRemoveMember(t *testing.T) {
	tests := []struct {
		name       string
		actor      TeamAccess
		targetRole int
		wantCode   errcode.Code // 0 表示放行
	}{
		{"管理员可移出普通成员", member(model.TeamRoleAdmin), model.TeamRoleMember, 0},
		{"所有者可移出管理员", member(model.TeamRoleOwner), model.TeamRoleAdmin, 0},
		{"管理员不能移出另一个管理员", member(model.TeamRoleAdmin), model.TeamRoleAdmin, errcode.OperationDenied},
		{"任何人都不能移出所有者", member(model.TeamRoleOwner), model.TeamRoleOwner, errcode.OperationDenied},
		{"普通成员无权移出别人", member(model.TeamRoleMember), model.TeamRoleMember, errcode.PermissionDenied},
		{"非成员无权移出别人", TeamAccess{UserID: 10}, model.TeamRoleMember, errcode.PermissionDenied},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := canRemoveMember(tt.actor, tt.targetRole)
			if tt.wantCode == 0 {
				if err != nil {
					t.Fatalf("canRemoveMember() error = %v, want nil", err)
				}
				return
			}
			requireAppErrorCode(t, err, tt.wantCode)
		})
	}
}

func TestCanQuitTeam(t *testing.T) {
	if err := canQuitTeam(member(model.TeamRoleMember)); err != nil {
		t.Fatalf("普通成员应能退队，got %v", err)
	}
	if err := canQuitTeam(member(model.TeamRoleAdmin)); err != nil {
		t.Fatalf("管理员应能退队，got %v", err)
	}
	// 所有者退队会让团队无人可管，必须先转让或解散
	requireAppErrorCode(t, canQuitTeam(member(model.TeamRoleOwner)), errcode.OperationDenied)
	requireAppErrorCode(t, canQuitTeam(TeamAccess{UserID: 10}), errcode.OperationDenied)
}

// TestHomeworkEditPolicy 覆盖用户明确选择的方案 B：
// 作业只有布置者本人能改，团队管理员乃至所有者都不行。
func TestHomeworkEditPolicy(t *testing.T) {
	hw := &model.Homework{ID: 1, CreatedBy: 10}

	t.Run("布置者本人可以改", func(t *testing.T) {
		if err := canEditHomework(TeamAccess{UserID: 10, IsMember: true, Role: model.TeamRoleAdmin}, hw); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})
	t.Run("团队管理员不能改别人布置的", func(t *testing.T) {
		actor := TeamAccess{UserID: 20, IsMember: true, Role: model.TeamRoleAdmin}
		requireAppErrorCode(t, canEditHomework(actor, hw), errcode.OperationDenied)
	})
	t.Run("团队所有者也不能改别人布置的", func(t *testing.T) {
		actor := TeamAccess{UserID: 30, IsMember: true, Role: model.TeamRoleOwner}
		requireAppErrorCode(t, canEditHomework(actor, hw), errcode.OperationDenied)
	})
	t.Run("站点超管例外", func(t *testing.T) {
		if err := canEditHomework(TeamAccess{UserID: 99, IsSiteAdmin: true}, hw); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})
}

// TestHomeworkDeletePolicy 删除比修改多给团队所有者兜底，
// 否则布置人离队后作业没人能清理。
func TestHomeworkDeletePolicy(t *testing.T) {
	hw := &model.Homework{ID: 1, CreatedBy: 10}

	t.Run("团队所有者可以删别人布置的", func(t *testing.T) {
		actor := TeamAccess{UserID: 30, IsMember: true, Role: model.TeamRoleOwner}
		if err := canDeleteHomework(actor, hw); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})
	t.Run("团队管理员仍不能删别人布置的", func(t *testing.T) {
		actor := TeamAccess{UserID: 20, IsMember: true, Role: model.TeamRoleAdmin}
		requireAppErrorCode(t, canDeleteHomework(actor, hw), errcode.OperationDenied)
	})
	t.Run("布置者本人可以删", func(t *testing.T) {
		actor := TeamAccess{UserID: 10, IsMember: true, Role: model.TeamRoleMember}
		if err := canDeleteHomework(actor, hw); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})
}
