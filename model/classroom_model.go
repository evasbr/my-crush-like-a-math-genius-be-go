package model

type CreateClassroomRequest struct {
	Name                   string  `json:"name" form:"name" validate:"required"`
	Description            *string `json:"description" form:"description"`
	CoverImg               *string `json:"cover_img" form:"cover_img"`
	WallpaperImg           *string `json:"wallpaper_img" form:"wallpaper_img"`
	IsExternalInviteEnable *bool   `json:"is_external_invite_enable" form:"is_external_invite_enable"`
}

type UpdateClassroomRequest struct {
	Name                   *string `json:"name" form:"name"`
	Description            *string `json:"description" form:"description"`
	CoverImg               *string `json:"cover_img" form:"cover_img"`
	WallpaperImg           *string `json:"wallpaper_img" form:"wallpaper_img"`
	IsExternalInviteEnable *bool   `json:"is_external_invite_enable" form:"is_external_invite_enable"`
	Status                 *string `json:"status" form:"status"`
}

type JoinClassroomRequest struct {
	Code string `json:"code" validate:"required"`
}

type ClassroomResponse struct {
	ID                     string  `json:"id"`
	Name                   string  `json:"name"`
	Description            *string `json:"description"`
	Codes                  *string `json:"codes,omitempty"`
	CoverImg               *string `json:"cover_img"`
	WallpaperImg           *string `json:"wallpaper_img"`
	IsExternalInviteEnable *bool   `json:"is_external_invite_enable,omitempty"`
	Status                 string  `json:"status"`
	CreatedAt              string  `json:"created_at"`
	UpdatedAt              string  `json:"updated_at"`
}

type ClassroomMemberResponse struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" form:"role" validate:"required,oneof=teacher student"`
}

type LeaderboardEntry struct {
	UserID            string  `json:"user_id"`
	Username          string  `json:"username"`
	FirstName         string  `json:"first_name"`
	LastName          *string `json:"last_name"`
	ProfilePictureURL *string `json:"profile_picture_url"`
	Score             int     `json:"score"`
	Rank              int     `json:"rank"`
}
