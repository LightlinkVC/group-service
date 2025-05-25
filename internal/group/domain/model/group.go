package model

type Group struct {
	ID        uint   `db:"id"`
	Name      string `db:"name"`
	CreatorID uint   `db:"creator_id"`
	TypeID    uint   `db:"type_id"`
}
