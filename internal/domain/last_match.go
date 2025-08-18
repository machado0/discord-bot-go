package domain

import "gorm.io/gorm"

type LastMatch struct {
    gorm.Model
    MatchId string `gorm:"column:match_id;not null" json:"match_id"`
    Puuid   string `gorm:"column:puuid;not null;uniqueIndex" json:"puuid"`
}

func (LastMatch) TableName() string {
    return "last_matches"
}