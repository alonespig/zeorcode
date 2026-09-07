package model

// RemoteAccount 远程 OJ 提交账号。
// Secret 存 AES 加密后的密码或 cookie/token(明文绝不落库)，
// json:"-" 保证即便误序列化也不会泄露。
type RemoteAccount struct {
	ID       int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	OJ       string `json:"oj" gorm:"type:varchar(20);not null;index"`
	AuthType string `json:"authType" gorm:"type:varchar(20);not null;default:'cookie'"` // password | cookie
	Username string `json:"username" gorm:"type:varchar(100);not null;default:''"`
	Secret   string `json:"-" gorm:"type:text;not null"`
	Enabled  bool   `json:"enabled" gorm:"not null;default:true"`
	Valid    bool   `json:"valid" gorm:"not null;default:false"`
	Busy     bool   `json:"-" gorm:"not null;default:false;index"` // 账号池占用标记：远程判题期间置 true

	CreatedAt int64 `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt int64 `json:"updated_at" gorm:"autoUpdateTime"`
}
