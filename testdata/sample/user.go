package sample

import "time"

// User はユーザー情報を表す構造体
type User struct {
	ID       int
	Name     string
	Email    string
	Profile  *Profile
	Posts    []Post
	Settings UserSettings
}

// Profile はユーザープロフィール情報
type Profile struct {
	Bio       string
	Avatar    string
	CreatedAt time.Time
}

// Post はユーザーの投稿
type Post struct {
	ID      int
	Title   string
	Content string
	Author  *User
}

// UserSettings はユーザー設定
type UserSettings struct {
	Theme         string
	Notifications bool
}

// UserService はユーザー関連のサービス
type UserService interface {
	GetUser(id int) (*User, error)
	CreateUser(user *User) error
	UpdateProfile(userID int, profile *Profile) error
}

// CreateUser は新しいユーザーを作成
func CreateUser(name, email string) *User {
	return &User{
		Name:  name,
		Email: email,
		Profile: &Profile{
			CreatedAt: time.Now(),
		},
		Settings: UserSettings{
			Theme:         "light",
			Notifications: true,
		},
	}
}

// GetUserPosts はユーザーの投稿一覧を取得
func GetUserPosts(user *User) []Post {
	return user.Posts
}

// UpdateUserProfile はユーザープロフィールを更新
func (u *User) UpdateProfile(bio, avatar string) {
	if u.Profile == nil {
		u.Profile = &Profile{}
	}
	u.Profile.Bio = bio
	u.Profile.Avatar = avatar
}

// AddPost はユーザーに投稿を追加
func (u *User) AddPost(title, content string) {
	post := Post{
		ID:      len(u.Posts) + 1,
		Title:   title,
		Content: content,
		Author:  u,
	}
	u.Posts = append(u.Posts, post)
}
