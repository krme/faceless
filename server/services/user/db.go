package user

import (
	"context"
	"fmt"
	"ht/server/database"
	"time"

	"github.com/google/uuid"
)

type UserDBHandlerFunctions interface {
	CreateTable(projectRid uuid.UUID) error
	DropTable(projectRid uuid.UUID) error
}

type UserDBHandler struct {
	db *database.Database
}

func newUserDBHandler(dbConnection *database.Database) *UserDBHandler {
	return &UserDBHandler{
		db: dbConnection,
	}
}

func (r UserDBHandler) CreateTable(projectRid uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.db.Instance.ExecContext(
		ctx,
		`CREATE EXTENSION IF NOT EXISTS pgcrypto;
		CREATE TABLE IF NOT EXISTS user (
			id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			rid UUID UNIQUE DEFAULT gen_random_uuid(),
			recording_1 ARRAY DEFAULT '[]'::ARRAY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	)
	if err != nil {
		return fmt.Errorf("error creating user table: %#v", err)
	}

	err = r.db.CreateIndex("user", "rid")
	if err != nil {
		return err
	}

	r.db.Logger.Println("created table user")
	return nil
}

func (r UserDBHandler) DropTable(projectRid uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `DROP TABLE IF EXISTS user`
	_, err := r.db.Instance.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error dropping user table: %#v", err)
	}

	r.db.Logger.Printf("dropped table user")
	return nil
}
