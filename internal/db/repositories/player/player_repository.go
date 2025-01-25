//go:generate mockgen -destination=mocks/mock_player_repository.go -package=mocks github.com/MyelinBots/pigeonbot-go/internal/db/repositories/player PlayerRepository
package player

import (
	"context"
	"errors"
	"github.com/MyelinBots/pigeonbot-go/internal/db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlayerRepository interface {
	GetPlayerByID(id string) (*Player, error)
	GetAllPlayers(ctx context.Context, network string, channel string) ([]*Player, error)
	UpsertPlayer(ctx context.Context, player *Player) error
}

type PlayerRepositoryImpl struct {
	db *db.DB
}

func NewPlayerRepository(db *db.DB) PlayerRepository {
	return &PlayerRepositoryImpl{
		db: db,
	}
}

func (r *PlayerRepositoryImpl) GetPlayerByID(id string) (*Player, error) {
	return nil, nil
}

func (r *PlayerRepositoryImpl) GetAllPlayers(ctx context.Context, network string, channel string) ([]*Player, error) {
	var players []*Player
	err := r.db.DB.Find(&players, "network = ? AND channel = ?", network, channel).Error
	if err != nil {
		return nil, err
	}

	return players, nil
}

func (r *PlayerRepositoryImpl) UpsertPlayer(ctx context.Context, player *Player) error {
	// find player by name
	var existing Player
	err := r.db.DB.Where("name = ? AND channel = ? AND network = ?", player.Name, player.Channel, player.Network).First(&existing).Error
	if err != nil {
		// if error type is gorm.ErrRecordNotFound, create new player
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// create uuid
			uuidString := uuid.New().String()
			player.ID = uuidString
			return r.db.DB.Create(player).Error
		}
		// if error is not gorm.ErrRecordNotFound, return error
		return err
	}
	existing.Points = player.Points
	existing.Count = player.Count
	// update player

	return r.db.DB.Save(existing).Error
}
