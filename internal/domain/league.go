package domain

type Player struct {
	ID        int64  `db:"id" json:"id"`
	GamerName string `db:"gamer_name" json:"gamer_name"`
	TagLine   string `db:"tag_line" json:"tag_line"`
	Puuid     string `db:"puuid" json:"puuid"`
	TeamID    int64  `db:"team_id" json:"team_id"`
}
