package identification

import (
	"context"
	"fmt"
	"ht/model"
	"ht/server/database"
	"log"
	"time"

	"github.com/google/uuid"
)

type IdentificationAttemptDBHandlerFunctions interface {
	CreateTable() error
	DropTable() error
	InsertIdentificationAttempt(identificationAttempt *model.IdentificationAttempt) (*model.IdentificationAttempt, error)
	UpdateIdentificationAttempt(identificationAttempt *model.IdentificationAttempt) (*model.IdentificationAttempt, error)
	SelectIdentificationAttempt(rid uuid.UUID) (*model.IdentificationAttempt, error)
	SelectLatestIdentificationAttemptByUserRID(userRid uuid.UUID) (*model.IdentificationAttempt, error)
	SelectAllIdentificationAttempts(lastId int, entries int) ([]*model.IdentificationAttempt, error)
	SelectAllIdentificationAttemptsBySearch(search string, lastId int, entries int) ([]*model.IdentificationAttempt, error)
}

type IdentificationAttemptDBHandler struct {
	db *database.Database
}

func newIdentificationAttemptDBHandler(dbConnection *database.Database) *IdentificationAttemptDBHandler {
	return &IdentificationAttemptDBHandler{
		db: dbConnection,
	}
}

func (r IdentificationAttemptDBHandler) CreateTable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.db.Instance.ExecContext(
		ctx,
		`CREATE EXTENSION IF NOT EXISTS vector;
		
		CREATE TABLE IF NOT EXISTS identification_attempt (
			id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			rid UUID UNIQUE DEFAULT gen_random_uuid(),
			user_rid UUID NOT NULL,
			recording BYTEA,
			recording_mfcc VECTOR(40),
			identified BOOLEAN,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
	)
	if err != nil {
		return fmt.Errorf("error creating identificationAttempt table: %v", err)
	}

	err = r.db.CreateIndex("identification_attempt", "rid")
	if err != nil {
		return err
	}

	r.db.Logger.Println("created table identificationAttempt")
	return nil
}

func (r IdentificationAttemptDBHandler) DropTable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `DROP TABLE IF EXISTS identification_attempt`
	_, err := r.db.Instance.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error dropping identificationAttempt table: %#v", err)
	}

	r.db.Logger.Printf("dropped table identificationAttempt")
	return nil
}

func (r IdentificationAttemptDBHandler) InsertIdentificationAttempt(identificationAttempt *model.IdentificationAttempt) (*model.IdentificationAttempt, error) {
	newIdentificationAttempt := &model.IdentificationAttempt{}

	row := r.db.Instance.QueryRow(
		`INSERT INTO identification_attempt (user_rid)
			VALUES ($1)
		RETURNING
			id,
			rid,
			user_rid,
			recording,
			identified,
			created_at,
			updated_at;`,
		identificationAttempt.UserRID,
	)

	err := row.Scan(
		&newIdentificationAttempt.ID,
		&newIdentificationAttempt.RID,
		&newIdentificationAttempt.UserRID,
		&newIdentificationAttempt.Recording,
		&newIdentificationAttempt.Identified,
		&newIdentificationAttempt.CreatedAt,
		&newIdentificationAttempt.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return newIdentificationAttempt, nil
}

func (r IdentificationAttemptDBHandler) UpdateIdentificationAttempt(identificationAttempt *model.IdentificationAttempt) (*model.IdentificationAttempt, error) {
	identificationAttemptUpdated := &model.IdentificationAttempt{}

	row := r.db.Instance.QueryRow(
		`UPDATE
			identification_attempt
		SET
			recording = $1,
            identified = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE
			rid = $2
		RETURNING
			id,
			rid,
			user_rid,
			recording,
			identified,
			created_at,
			updated_at;`,
		identificationAttempt.Recording,
		identificationAttempt.Identified,
		identificationAttempt.RID,
	)

	err := row.Scan(
		&identificationAttemptUpdated.ID,
		&identificationAttemptUpdated.RID,
		&identificationAttemptUpdated.UserRID,
		&identificationAttemptUpdated.Recording,
		&identificationAttemptUpdated.Identified,
		&identificationAttemptUpdated.CreatedAt,
		&identificationAttemptUpdated.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return identificationAttemptUpdated, err
}

func (r IdentificationAttemptDBHandler) DeleteIdentificationAttempt(rid uuid.UUID) error {
	_, err := r.db.Instance.Exec(
		`DELETE FROM identification_attempt
		WHERE rid = $1`,
		rid,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r IdentificationAttemptDBHandler) SelectIdentificationAttempt(rid uuid.UUID) (*model.IdentificationAttempt, error) {
	identificationAttempt := &model.IdentificationAttempt{}

	row := r.db.Instance.QueryRow(
		`SELECT
			id,
			rid,
			user_rid,
			recording,
			identified,
			created_at,
			updated_at
		FROM
			identification_attempt
		WHERE
			rid = $1`,
		rid,
	)
	err := row.Scan(
		&identificationAttempt.ID,
		&identificationAttempt.RID,
		&identificationAttempt.UserRID,
		&identificationAttempt.Recording,
		&identificationAttempt.Identified,
		&identificationAttempt.CreatedAt,
		&identificationAttempt.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return identificationAttempt, nil
}

func (r IdentificationAttemptDBHandler) SelectLatestIdentificationAttemptByUserRID(userRid uuid.UUID) (*model.IdentificationAttempt, error) {
	identificationAttempt := &model.IdentificationAttempt{}

	row := r.db.Instance.QueryRow(
		`SELECT
			id,
			rid,
			user_rid,
			recording,
			identified,
			created_at,
			updated_at
		FROM
			identification_attempt
		WHERE
			user_rid = $1
		ORDER BY
			created_at DESC
		LIMIT 1`,
		userRid,
	)
	err := row.Scan(
		&identificationAttempt.ID,
		&identificationAttempt.RID,
		&identificationAttempt.UserRID,
		&identificationAttempt.Recording,
		&identificationAttempt.Identified,
		&identificationAttempt.CreatedAt,
		&identificationAttempt.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return identificationAttempt, nil
}

func (r IdentificationAttemptDBHandler) SelectAllIdentificationAttempts(lastId int, entries int) ([]*model.IdentificationAttempt, error) {
	var identificationAttempts []*model.IdentificationAttempt

	rows, err := r.db.Instance.Query(
		`SELECT
			id,
			rid,
			user_rid,
			recording,
			identified,
			created_at,
			updated_at
		FROM
			identification_attempt
		WHERE (0 = $1
			OR created_at < (
				SELECT
					u.created_at
				FROM
					identification_attempt AS u
				WHERE
					u.id = $1))
		ORDER BY
			created_at DESC
		LIMIT $2`,
		lastId,
		entries,
	)
	if err != nil {
		return []*model.IdentificationAttempt{}, err
	}

	defer rows.Close()

	for rows.Next() {
		identificationAttempt := &model.IdentificationAttempt{}
		err := rows.Scan(
			&identificationAttempt.ID,
			&identificationAttempt.RID,
			&identificationAttempt.UserRID,
			&identificationAttempt.Recording,
			&identificationAttempt.Identified,
			&identificationAttempt.CreatedAt,
			&identificationAttempt.UpdatedAt,
		)
		if err != nil {
			return []*model.IdentificationAttempt{}, err
		}

		identificationAttempts = append(identificationAttempts, identificationAttempt)
	}

	return identificationAttempts, nil
}

func (r IdentificationAttemptDBHandler) SelectAllIdentificationAttemptsBySearch(search string, lastId int, entries int) ([]*model.IdentificationAttempt, error) {
	var identificationAttempts []*model.IdentificationAttempt

	log.Printf("search: %v", search)

	rows, err := r.db.Instance.Query(
		`SELECT
			id,
			rid,
			user_rid,
			recording,
			identified,
			created_at,
			updated_at
		FROM identification_attempt 
		WHERE (key ILIKE '%' || $1 || '%'
				OR name ILIKE '%' || $1 || '%'
				OR description ILIKE '%' || $1 || '%')
			AND (0 = $2
				OR created_at < (
					SELECT
						u.created_at
					FROM
						identification_attempt AS u
					WHERE
						u._id = $2))
		ORDER BY
			created_at DESC
		LIMIT $3`,
		search,
		lastId,
		entries,
	)
	if err != nil {
		return []*model.IdentificationAttempt{}, err
	}

	defer rows.Close()

	for rows.Next() {
		identificationAttempt := &model.IdentificationAttempt{}
		err := rows.Scan(
			&identificationAttempt.ID,
			&identificationAttempt.RID,
			&identificationAttempt.UserRID,
			&identificationAttempt.Recording,
			&identificationAttempt.Identified,
			&identificationAttempt.CreatedAt,
			&identificationAttempt.UpdatedAt,
		)
		if err != nil {
			return []*model.IdentificationAttempt{}, err
		}

		identificationAttempts = append(identificationAttempts, identificationAttempt)
	}

	return identificationAttempts, err
}
