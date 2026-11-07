// Package authz 定义后台权限点及其分组目录。
//
// 权限点是代码里的固定常量：角色表/权限表只存这些字符串，
// 前端勾选界面也直接使用目录接口返回的分组，避免两端各写一份列表。
package authz

// 权限点 key。命名规范：模块.动作。
const (
	DashboardRead = "dashboard.read"

	ProblemRead    = "problem.read"
	ProblemWrite   = "problem.write"
	ProblemDelete  = "problem.delete"
	ProblemArchive = "problem.archive"

	ProblemSetRead   = "problemset.read"
	ProblemSetWrite  = "problemset.write"
	ProblemSetDelete = "problemset.delete"

	ContestRead    = "contest.read"
	ContestWrite   = "contest.write"
	ContestDelete  = "contest.delete"
	ContestRejudge = "contest.rejudge"

	TagRead  = "tag.read"
	TagWrite = "tag.write"

	PostReview = "post.review"

	TeamManage = "team.manage"

	SubmissionRead    = "submission.read"
	SubmissionRejudge = "submission.rejudge"

	JudgeNodeRead       = "judge.node.read"
	JudgeNodeManage     = "judge.node.manage"
	JudgeLanguageRead   = "judge.language.read"
	JudgeLanguageManage = "judge.language.manage"
	JudgeRemoteRead     = "judge.remote.read"
	JudgeRemoteManage   = "judge.remote.manage"

	UserRead   = "user.read"
	UserManage = "user.manage"
	UserImport = "user.import"

	NotificationSend = "notification.send"

	AuditRead = "audit.read"
)

// Permission 是目录中的一个权限点，Label 用于前端展示。
type Permission struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// Group 是一个权限分组，对应管理后台的一个功能模块。
type Group struct {
	Key         string       `json:"key"`
	Label       string       `json:"label"`
	Permissions []Permission `json:"permissions"`
}

// Catalog 是权限分组目录，顺序即前端展示顺序。
var Catalog = []Group{
	{Key: "dashboard", Label: "工作台", Permissions: []Permission{
		{Key: DashboardRead, Label: "查看工作台"},
	}},
	{Key: "problem", Label: "题目管理", Permissions: []Permission{
		{Key: ProblemRead, Label: "查看题目"},
		{Key: ProblemWrite, Label: "编辑题目"},
		{Key: ProblemDelete, Label: "删除题目"},
		{Key: ProblemArchive, Label: "归档导入导出"},
	}},
	{Key: "problemset", Label: "题单管理", Permissions: []Permission{
		{Key: ProblemSetRead, Label: "查看题单"},
		{Key: ProblemSetWrite, Label: "编辑题单"},
		{Key: ProblemSetDelete, Label: "删除题单"},
	}},
	{Key: "contest", Label: "比赛管理", Permissions: []Permission{
		{Key: ContestRead, Label: "查看比赛"},
		{Key: ContestWrite, Label: "编辑比赛"},
		{Key: ContestDelete, Label: "删除比赛"},
		{Key: ContestRejudge, Label: "重新评测"},
	}},
	{Key: "tag", Label: "标签管理", Permissions: []Permission{
		{Key: TagRead, Label: "查看标签"},
		{Key: TagWrite, Label: "编辑标签"},
	}},
	{Key: "post", Label: "帖子审核", Permissions: []Permission{
		{Key: PostReview, Label: "审核帖子"},
	}},
	{Key: "team", Label: "团队管理", Permissions: []Permission{
		{Key: TeamManage, Label: "管理全站团队"},
	}},
	{Key: "submission", Label: "提交记录", Permissions: []Permission{
		{Key: SubmissionRead, Label: "查看提交"},
		{Key: SubmissionRejudge, Label: "重新评测"},
	}},
	{Key: "judge-node", Label: "评测机", Permissions: []Permission{
		{Key: JudgeNodeRead, Label: "查看节点"},
		{Key: JudgeNodeManage, Label: "管理节点"},
	}},
	{Key: "judge-language", Label: "编程语言", Permissions: []Permission{
		{Key: JudgeLanguageRead, Label: "查看语言"},
		{Key: JudgeLanguageManage, Label: "管理语言"},
	}},
	{Key: "judge-remote", Label: "远程账号", Permissions: []Permission{
		{Key: JudgeRemoteRead, Label: "查看账号"},
		{Key: JudgeRemoteManage, Label: "管理账号"},
	}},
	{Key: "user", Label: "用户管理", Permissions: []Permission{
		{Key: UserRead, Label: "查看用户"},
		{Key: UserManage, Label: "编辑用户"},
		{Key: UserImport, Label: "批量导入"},
	}},
	{Key: "notification", Label: "系统通知", Permissions: []Permission{
		{Key: NotificationSend, Label: "发送通知"},
	}},
	{Key: "audit", Label: "操作日志", Permissions: []Permission{
		{Key: AuditRead, Label: "查看日志"},
	}},
}

var validKeys = func() map[string]struct{} {
	keys := make(map[string]struct{})
	for _, group := range Catalog {
		for _, perm := range group.Permissions {
			keys[perm.Key] = struct{}{}
		}
	}
	return keys
}()

var allKeys = func() []string {
	keys := make([]string, 0, len(validKeys))
	for _, group := range Catalog {
		for _, perm := range group.Permissions {
			keys = append(keys, perm.Key)
		}
	}
	return keys
}()

// IsValid 判断权限点是否存在于目录中。
func IsValid(key string) bool {
	_, ok := validKeys[key]
	return ok
}

// All 返回全部权限点，供超级管理员使用。
func All() []string {
	out := make([]string, len(allKeys))
	copy(out, allKeys)
	return out
}
