package service

import (
	"strings"
	"testing"

	"zoj/internal/common/errcode"
	"zoj/internal/model"
)

func TestUsernamePattern(t *testing.T) {
	tests := []struct {
		username string
		valid    bool
	}{
		{username: "student_01", valid: true},
		{username: "A2", valid: true},
		{username: "a", valid: false},
		{username: "name@example.com", valid: false},
		{username: "中文昵称", valid: false},
	}
	for _, tt := range tests {
		if got := usernamePattern.MatchString(tt.username); got != tt.valid {
			t.Errorf("usernamePattern.MatchString(%q) = %v, want %v", tt.username, got, tt.valid)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantError bool
	}{
		{name: "普通密码", password: "student123", wantError: false},
		{name: "少于六位", password: "12345", wantError: true},
		{name: "中文按字符计最小长度", password: "六个汉字密码", wantError: false},
		{name: "超过 bcrypt 字节上限", password: strings.Repeat("密", 25), wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)
			if (err != nil) != tt.wantError {
				t.Fatalf("validatePassword() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestLoginResponse(t *testing.T) {
	tests := []struct {
		name       string
		avatar     string
		wantAvatar string
	}{
		{name: "保留用户头像", avatar: "/uploads/avatar.png", wantAvatar: "/uploads/avatar.png"},
		{name: "空头像使用默认值", wantAvatar: defaultAvatar},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &model.User{
				UID:      12345678,
				Username: "student_01",
				Role:     model.RoleAdmin,
				Avatar:   tt.avatar,
			}
			got := loginResponse(user)
			if got.User.ID != user.UID || got.User.Username != user.Username || got.User.Role != user.Role {
				t.Fatalf("loginResponse() user = %+v, want uid/username/role from model", got.User)
			}
			if got.Avatar != tt.wantAvatar {
				t.Fatalf("loginResponse() avatar = %q, want %q", got.Avatar, tt.wantAvatar)
			}
		})
	}
}

func TestSchoolIdentityValueHelpers(t *testing.T) {
	nameTests := []struct {
		name     string
		realName string
		username string
		want     string
	}{
		{name: "未填写真实姓名时使用用户名", username: "student_01", want: "student_01"},
		{name: "空白真实姓名时使用用户名", realName: "  ", username: "student_01", want: "student_01"},
		{name: "保留并清理已填写的真实姓名", realName: " 张三 ", username: "student_01", want: "张三"},
	}
	for _, tt := range nameTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := registrationRealName(tt.realName, tt.username); got != tt.want {
				t.Fatalf("registrationRealName(%q, %q) = %q, want %q", tt.realName, tt.username, got, tt.want)
			}
		})
	}

	if got := studentNoPtr(""); got != nil {
		t.Fatalf("studentNoPtr(empty) = %v, want nil", got)
	}
	got := studentNoPtr("20260001")
	if got == nil || *got != "20260001" {
		t.Fatalf("studentNoPtr(value) = %v, want 20260001", got)
	}
	if value := studentNoValue(nil); value != "" {
		t.Fatalf("studentNoValue(nil) = %q, want empty", value)
	}
	if email := normalizeEmail(" Student@Example.COM "); email != "student@example.com" {
		t.Fatalf("normalizeEmail() = %q, want student@example.com", email)
	}
	if got := emailPtr(""); got != nil {
		t.Fatalf("emailPtr(empty) = %v, want nil", got)
	}
	email := emailPtr("student@example.com")
	if got := emailValue(email); got != "student@example.com" {
		t.Fatalf("emailValue() = %q, want student@example.com", got)
	}
	if got := emailValue(nil); got != "" {
		t.Fatalf("emailValue(nil) = %q, want empty", got)
	}
}

func TestChangeRolePolicy(t *testing.T) {
	tests := []struct {
		name     string
		actorID  int64
		targetID int64
		newRole  int
		wantCode errcode.Code // 0 表示期望放行
	}{
		{
			name:     "不能修改自己的角色",
			actorID:  1,
			targetID: 1,
			newRole:  model.RoleNormal,
			wantCode: errcode.OperationDenied,
		},
		{
			name:     "可以把普通用户升为管理员",
			actorID:  1,
			targetID: 2,
			newRole:  model.RoleAdmin,
		},
		{
			name:     "可以把其他管理员降为普通用户",
			actorID:  1,
			targetID: 2,
			newRole:  model.RoleNormal,
		},
		{
			name:     "非法角色值被拒绝",
			actorID:  1,
			targetID: 2,
			newRole:  2,
			wantCode: errcode.InvalidParams,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := changeRolePolicy(tt.actorID, tt.targetID, tt.newRole)
			if tt.wantCode == 0 {
				if err != nil {
					t.Fatalf("changeRolePolicy() error = %v, want nil", err)
				}
				return
			}
			requireAppErrorCode(t, err, tt.wantCode)
		})
	}
}
