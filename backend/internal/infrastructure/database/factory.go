package database

import (
	"context"
	"videocall/internal/domain/repositories/db"
	"videocall/internal/domain/repositories/mem"

	"videocall/internal/domain/repositories"
	"videocall/internal/infrastructure/config"
)

const TypeMaria = "mariadb"

type StorageFactory struct {
	storageType string
	db          *Database
}

func NewStorageFactory(cfg *config.Storage) (*StorageFactory, error) {
	factory := &StorageFactory{
		storageType: cfg.Type,
	}

	if cfg.Type == TypeMaria {
		db, err := NewDatabase(cfg)
		if err != nil {
			return nil, err
		}
		factory.db = db
	}

	return factory, nil
}

// CreateRoomRepository creates a room repository based on storage type
func (f *StorageFactory) CreateRoomRepository(ctx context.Context) repositories.RoomRepositoryInterface {
	if f.storageType == TypeMaria {
		return db.NewMariaDBRoomRepository(f.db.GetDB())
	}

	// Default to in-memory storage
	return mem.New()
}

// CreateUserRepository creates a user repository based on storage type
func (f *StorageFactory) CreateUserRepository() repositories.UserRepositoryInterface {
	if f.storageType == TypeMaria {
		return db.NewMariaDBUserRepository(f.db.GetDB())
	}

	// Default to in-memory storage
	return mem.NewUserRepository()
}

// CreateRefreshTokenRepository creates a refresh token repository based on storage type
func (f *StorageFactory) CreateRefreshTokenRepository(ctx context.Context) repositories.RefreshTokenRepositoryInterface {
	if f.storageType == TypeMaria {
		return db.NewMariaDBRefreshTokenRepository(f.db.GetDB())
	}

	// Default to in-memory storage
	return mem.NewRefreshTokenRepository(ctx)
}

// Close closes the database connection if using MariaDB
func (f *StorageFactory) Close() error {
	if f.db != nil {
		return f.db.Close()
	}
	return nil
}
