package db

import "nextChatServer/internal/model"

func AutoMigrate() error {
	return GetDB().AutoMigrate(
		&model.User{},
		&model.Profile{},
		&model.Message{},
	)
}
