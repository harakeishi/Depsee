package pkg1

import (
	"time"

	"github.com/harakeishi/depsee/testdata/multi-package/pkg2"
)

// User はユーザー情報を表す構造体
type User struct {
	ID       int
	Name     string
	Email    string
	Profile  *pkg2.Profile
	Settings pkg2.Settings
}

// UserService はユーザー関連のサービス
type UserService interface {
	GetUser(id int) (*User, error)
	CreateUser(user *User) error
}

// CreateUser は新しいユーザーを作成
func CreateUser(name, email string) *User {
	return &User{
		Name:  name,
		Email: email,
		Profile: &pkg2.Profile{
			CreatedAt: time.Now(),
		},
		Settings: pkg2.Settings{
			Theme:         "light",
			Notifications: true,
		},
	}
}

// GetUserProfile はユーザープロフィールを取得
func GetUserProfile(user *User) *pkg2.Profile {
	return user.Profile
}
