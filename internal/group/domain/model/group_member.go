package model

type GroupMember struct {
	UserID  uint `db:"user_id"`
	GroupID uint `db:"group_id"`
	RoleID  uint `db:"role_id"`
}
