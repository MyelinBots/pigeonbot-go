package player

import (
	"time"

	"gorm.io/gorm"
)

type Player struct {
	gorm.Model `json:"-"`
	ID         string    `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name       string    `gorm:"column:name;type:text;not null" json:"name"`
	Points     int       `gorm:"column:points;type:int;not null" json:"points"`
	Count      int       `gorm:"column:count;type:int;not null" json:"count"`
	Network    string    `gorm:"column:network;type:text;not null" json:"network"`
	Channel    string    `gorm:"column:channel;type:text;not null" json:"channel"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updated_at"`
	Eggs       int       `gorm:"column:eggs;type:int;not null;default:0" json:"eggs"`
	RareEggs   int       `gorm:"column:rare_eggs;type:int;not null;default:0" json:"rare_eggs"`
}

// set table name
func (Player) TableName() string {
	return "player"
}
