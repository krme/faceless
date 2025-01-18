package user

import (
	"context"
	"fmt"
	"ht/model"
	"ht/server/database"
	"log"
	"time"

	"github.com/google/uuid"
)

type UserDBHandlerFunctions interface {
	CreateTable() error
	DropTable() error
	InsertUser(user *model.User) (*model.User, error)
}

type UserDBHandler struct {
	db *database.Database
}

func newUserDBHandler(dbConnection *database.Database) *UserDBHandler {
	return &UserDBHandler{
		db: dbConnection,
	}
}

func (r UserDBHandler) CreateTable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.db.Instance.ExecContext(
		ctx,
		`CREATE EXTENSION IF NOT EXISTS vector;
        
        CREATE TABLE IF NOT EXISTS "user" (
            id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
            rid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
            recording_1 BYTEA,
            recording_2 BYTEA,
            recording_3 BYTEA,
			recording_1_normalised BYTEA,
            recording_2_normalised BYTEA,
            recording_3_normalised BYTEA,
			recording_1_mfcc VECTOR(40),
            recording_2_mfcc VECTOR(40),
            recording_3_mfcc VECTOR(40),
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );`,
	)
	if err != nil {
		return fmt.Errorf("error creating user table: %v", err)
	}

	err = r.db.CreateIndex("user", "rid")
	if err != nil {
		return err
	}

	r.db.Logger.Println("created table user")
	return nil
}

func (r UserDBHandler) DropTable() error {
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

func (r UserDBHandler) InsertUser(user *model.User) (*model.User, error) {
	newData := &model.User{}

	row := r.db.Instance.QueryRow(
		`INSERT INTO user ( id, rid, recording_1, recording_2, recording_3, recording_1_normalised, recording_2_normalised, recording_3_normalised, recording_1_mfcc, recording_2_mfcc, recording_3_mfcc )
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING
			id, rid , created_at, updated_at`,
		user.RID,
	)

	err := row.Scan(
		&newData.ID,
		&newData.RID,
		&newData.Recording1,
		&newData.Recording2,
		&newData.Recording3,
		&newData.Recording1Normalised,
		&newData.Recording2Normalised,
		&newData.Recording3Normalised,
		&newData.Recording1Mfcc,
		&newData.Recording2Mfcc,
		&newData.Recording3Mfcc,
		&newData.CreatedAt,
		&newData.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return newData, nil
}

func (r UserDBHandler) UpdateUser(user *model.User) error {
	_, err := r.db.Instance.Exec(
		`UPDATE
            "user"
        SET
            recording_1 = $1,
            recording_2 = $2,
            recording_3 = $3,
			recording_1_normalised = $4,
			recording_2_normalised = $5,
			recording_3_normalised = $6,
			recording_1_mfcc = $7,
			recording_2_mfcc = $8,
			recording_3_mfcc = $9,
            updated_at = CURRENT_TIMESTAMP
        WHERE
            rid = $10`,
		user.Recording1,
		user.Recording2,
		user.Recording3,
		user.Recording1Normalised,
		user.Recording2Normalised,
		user.Recording3Normalised,
		user.Recording1Mfcc,
		user.Recording2Mfcc,
		user.Recording3Mfcc,
		user.RID,
	)

	return err
}

func (r UserDBHandler) DeleteData(projectRid uuid.UUID, datamodelRid uuid.UUID, rid uuid.UUID) error {
	_, err := r.db.Instance.Exec(
		`DELETE FROM "user"
		WHERE rid = $1`,
		rid,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r UserDBHandler) SelectData(projectRid uuid.UUID, datamodelRid uuid.UUID, rid uuid.UUID) (*model.User, error) {
	data := &model.User{}

	row := r.db.Instance.QueryRow(
		`SELECT
			id,
			rid,
			recording_1,
			recording_2,
			recording_3,
			recording_1_normalised,
			recording_2_normalised,
			recording_3_normalised,
			recording_1_mfcc,
			recording_2_mfcc,
			recording_3_mfcc,
			created_at,
			updated_at
		FROM
			"user"
		WHERE
			rid = $1`,
		rid,
	)
	err := row.Scan(
		&data.ID,
		&data.RID,
		&data.Recording1,
		&data.Recording2,
		&data.Recording3,
		&data.Recording1Normalised,
		&data.Recording2Normalised,
		&data.Recording3Normalised,
		&data.Recording1Mfcc,
		&data.Recording2Mfcc,
		&data.Recording3Mfcc,
		&data.CreatedAt,
		&data.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r UserDBHandler) SelectAllData(projectRid uuid.UUID, datamodelRid uuid.UUID, lastId int, entries int) ([]*model.User, error) {
	var datas []*model.User

	rows, err := r.db.Instance.Query(
		`SELECT
			id,
			rid,
			recording_1,
			recording_2,
			recording_3,
			recording_1_normalised,
			recording_2_normalised,
			recording_3_normalised,
			recording_1_mfcc,
			recording_2_mfcc,
			recording_3_mfcc,
			created_at,
			updated_at
		FROM
			"user"
		WHERE (0 = $1
			OR created_at < (
				SELECT
					u.created_at
				FROM
					"user" AS u
				WHERE
					u.id = $1))
		ORDER BY
			created_at DESC
		LIMIT $2`,
		lastId,
		entries,
	)
	if err != nil {
		return []*model.User{}, err
	}

	defer rows.Close()

	for rows.Next() {
		data := &model.User{}
		err := rows.Scan(
			&data.RID,
			&data.Recording1,
			&data.Recording2,
			&data.Recording3,
			&data.Recording1Normalised,
			&data.Recording2Normalised,
			&data.Recording3Normalised,
			&data.Recording1Mfcc,
			&data.Recording2Mfcc,
			&data.Recording3Mfcc,
			&data.CreatedAt,
			&data.UpdatedAt,
		)
		if err != nil {
			return []*model.User{}, err
		}

		datas = append(datas, data)
	}

	return datas, nil
}

func (r UserDBHandler) SelectAllDataBySearch(projectRid uuid.UUID, datamodelRid uuid.UUID, search string, lastId int, entries int) ([]*model.User, error) {
	var datas []*model.User

	log.Printf("search: %v", search)

	rows, err := r.db.Instance.Query(
		`SELECT
			id,
			rid,
			recording_1,
			recording_2,
			recording_3,
			recording_1_normalised,
			recording_2_normalised,
			recording_3_normalised,
			recording_1_mfcc,
			recording_2_mfcc,
			recording_3_mfcc,
			created_at,
			updated_at
		FROM "user" 
		WHERE (key ILIKE '%' || $1 || '%'
				OR name ILIKE '%' || $1 || '%'
				OR description ILIKE '%' || $1 || '%')
			AND (0 = $2
				OR created_at < (
					SELECT
						u.created_at
					FROM
						"user" AS u
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
		return []*model.User{}, err
	}

	defer rows.Close()

	for rows.Next() {
		data := &model.User{}
		err := rows.Scan(
			&data.RID,
			&data.Recording1,
			&data.Recording2,
			&data.Recording3,
			&data.Recording1Normalised,
			&data.Recording2Normalised,
			&data.Recording3Normalised,
			&data.Recording1Mfcc,
			&data.Recording2Mfcc,
			&data.Recording3Mfcc,
			&data.CreatedAt,
			&data.UpdatedAt,
		)
		if err != nil {
			return []*model.User{}, err
		}

		datas = append(datas, data)
	}

	return datas, err
}
