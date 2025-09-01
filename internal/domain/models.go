package domain

import "gorm.io/gorm"

type Anger struct {
	gorm.Model
	User string `gorm:"column:user;not null" json:"user"`
}

func (Anger) TableName() string {
	return "tilted"
}

type Birthday struct {
	gorm.Model
	User     string
	Day      int
	Month    int
	ServerID uint
	Server   Server
}

func (Birthday) TableName() string {
    return "birthdays"
}

type Server struct {
	gorm.Model
	Guild           string
	BirthdayChannel string
}

func (Server) TableName() string {
    return "servers"
}

type LastMatch struct {
    gorm.Model
    MatchId string `gorm:"column:match_id;not null" json:"match_id"`
    Puuid   string `gorm:"column:puuid;not null;uniqueIndex" json:"puuid"`
}

func (LastMatch) TableName() string {
    return "last_matches"
}

type Player struct {
	ID        int64  `gorm:"id" json:"id"`
	GamerName string `gorm:"gamer_name" json:"gamer_name"`
	TagLine   string `gorm:"tag_line" json:"tag_line"`
	Puuid     string `gorm:"puuid" json:"puuid"`
	TeamID    int64  `gorm:"team_id" json:"team_id"`
}

type Bet struct {
	gorm.Model
	WillWin bool 	`gorm:"will_win" json:"will_win"`
	User   	string 	`gorm:"user;not null;uniqueIndex" json:"user"`
	Result 	bool	`gorm:"result" json:"result"`
	MatchId string	`gorm:"match_id" json:"match_id"`
}

func (Bet) TableName() string {
    return "bets"
}

type BetResults struct {
	gorm.Model
	User   	string 	`gorm:"user;not null;uniqueIndex" json:"user"`
	Points 	int64	`gorm:"result" json:"result"`
}

func (BetResults) TableName() string {
    return "bet_results"
}

type OpenAICompletionRequest struct {
	Model    string        `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAICompletionResponse struct {
	Choices []struct {
		Message OpenAIMessage `json:"message"`
	} `json:"choices"`
}

type OpenAIEmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type OpenAIEmbeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}