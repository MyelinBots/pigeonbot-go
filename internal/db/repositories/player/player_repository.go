//go:generate mockgen -destination=mocks/mock_player_repository.go -package=mocks github.com/MyelinBots/pigeonbot-go/internal/db/repositories/player PlayerRepository
package player

import (
	"context"
	"errors"
	"strings"

	"github.com/MyelinBots/pigeonbot-go/internal/db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// canonicalName ensures consistent lowercase name storage
func canonicalName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

type PlayerRepository interface {
	GetPlayerByID(id string) (*Player, error)
	GetAllPlayers(ctx context.Context, network string, channel string) ([]*Player, error)
	UpsertPlayer(ctx context.Context, player *Player) error
	TopByPoints(ctx context.Context, network, channel string, limit int) ([]*Player, error)
	AddEggs(ctx context.Context, network, channel, name string, delta int) (newTotal int, err error)
	GetEggs(ctx context.Context, network, channel, name string) (total int, err error)
	AddRareEggs(ctx context.Context, network, channel, name string, delta int) (newTotal int, err error)
	GetRareEggs(ctx context.Context, network, channel, name string) (total int, err error)
}

type PlayerRepositoryImpl struct {
	db *db.DB
}

type InventoryRepo interface {
	AddEggs(ctx context.Context, userID string, delta int) (newTotal int, err error)
	GetEggs(ctx context.Context, userID string) (total int, err error)
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
	// Canonicalize name for consistent storage
	player.Name = canonicalName(player.Name)

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

	return r.db.DB.Save(&existing).Error
}

// TopByPoints returns the top N players by points (and count as tiebreaker).
func (r *PlayerRepositoryImpl) TopByPoints(ctx context.Context, network, channel string, limit int) ([]*Player, error) {
	if limit <= 0 {
		limit = 5
	}
	if limit > 50 {
		limit = 50
	}

	var players []*Player
	err := r.db.DB.WithContext(ctx).
		Where("network = ? AND channel = ?", network, channel).
		Order("points DESC").
		Order("count DESC").
		Order("name ASC").
		Limit(limit).
		Find(&players).Error
	if err != nil {
		return nil, err
	}

	return players, nil
}

func (r *PlayerRepositoryImpl) GetEggs(ctx context.Context, network, channel, name string) (int, error) {
	name = canonicalName(name)

	var p Player
	err := r.db.DB.WithContext(ctx).
		Where("name = ? AND channel = ? AND network = ?", name, channel, network).
		First(&p).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return p.Eggs, nil
}

func (r *PlayerRepositoryImpl) AddEggs(ctx context.Context, network, channel, name string, delta int) (int, error) {
	name = canonicalName(name)

	if delta <= 0 {
		return r.GetEggs(ctx, network, channel, name)
	}

	// Ensure player exists (create minimal row if not found)
	var p Player
	err := r.db.DB.WithContext(ctx).
		Where("name = ? AND channel = ? AND network = ?", name, channel, network).
		First(&p).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			p = Player{
				ID:      uuid.New().String(),
				Name:    name,
				Channel: channel,
				Network: network,
				Eggs:    0,
			}
			if err := r.db.DB.WithContext(ctx).Create(&p).Error; err != nil {
				return 0, err
			}
		} else {
			return 0, err
		}
	}

	// Atomic increment (no race conditions)
	if err := r.db.DB.WithContext(ctx).
		Model(&Player{}).
		Where("name = ? AND channel = ? AND network = ?", name, channel, network).
		UpdateColumn("eggs", gorm.Expr("eggs + ?", delta)).Error; err != nil {
		return 0, err
	}

	return r.GetEggs(ctx, network, channel, name)
}

func (r *PlayerRepositoryImpl) GetRareEggs(ctx context.Context, network, channel, name string) (int, error) {
	name = canonicalName(name)

	var p Player
	err := r.db.DB.WithContext(ctx).
		Select("rare_eggs").
		Where("network = ? AND channel = ? AND name = ?", network, channel, name).
		First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	return p.RareEggs, err
}

func (r *PlayerRepositoryImpl) AddRareEggs(ctx context.Context, network, channel, name string, delta int) (int, error) {
	name = canonicalName(name)

	if delta <= 0 {
		return r.GetRareEggs(ctx, network, channel, name)
	}

	// Ensure player exists (reuse your Upsert logic if needed)
	// If you already create player rows elsewhere, you can skip this part.

	err := r.db.DB.WithContext(ctx).
		Model(&Player{}).
		Where("network = ? AND channel = ? AND name = ?", network, channel, name).
		UpdateColumn("rare_eggs", gorm.Expr("rare_eggs + ?", delta)).
		Error
	if err != nil {
		return 0, err
	}

	return r.GetRareEggs(ctx, network, channel, name)
}
