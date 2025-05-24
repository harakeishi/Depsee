package pkg2

import "time"

// Profile はユーザープロフィール情報
type Profile struct {
	Bio       string
	Avatar    string
	CreatedAt time.Time
}

// Settings はユーザー設定
type Settings struct {
	Theme         string
	Notifications bool
}

// UpdateProfile はプロフィールを更新
func (p *Profile) UpdateProfile(bio, avatar string) {
	p.Bio = bio
	p.Avatar = avatar
}

// GetDisplayName は表示名を取得
func GetDisplayName(profile *Profile) string {
	if profile.Bio != "" {
		return profile.Bio
	}
	return "No bio"
}
