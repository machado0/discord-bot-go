package domain

import "gorm.io/gorm"

type Birthday struct {
	gorm.Model
	User     string
	Day      int
	Month    int
	ServerID uint
	Server   Server
}

type Server struct {
	gorm.Model
	Guild           string
	BirthdayChannel string
}

func (Server) TableName() string {
    return "servers"
}