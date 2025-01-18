package user

import (
	"context"
	"fmt"
	"ht/model"
	"ht/server/database"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
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
		`CREATE EXTENSION IF NOT EXISTS pgcrypto;
		CREATE TABLE IF NOT EXISTS user (
			id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			rid UUID UNIQUE NOT NULL,
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
		`INSERT INTO user ( TODO add columns )
			VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id, rid, TODO add columns, created_at, updated_at`,
		user.RID,
	)

	err := row.Scan(
		&newData.ID,
		&newData.RID,
		// TODO add User columns
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
			user
		SET
			owner_rid = $1,
			parent_id = NULLIF ($2, 0),
			parent_rid = $3,
			parent_column = $4,
			details = $5,
			updated_at = CURRENT_TIMESTAMP
		WHERE
			rid = $6`,
		user.OwnerRID,
		user.ParentID,
		user.ParentRID,
		user.ParentColumn,
		user.Details,
		user.RID,
	)

	return err
}

func (r DataDBHandler) DeleteData(projectRid uuid.UUID, datamodelRid uuid.UUID, rid uuid.UUID) error {
	tableNameQuoted := pq.QuoteIdentifier(r.tableNameByProjectRidAndDatamodelRid(projectRid, datamodelRid))
	_, err := r.db.Instance.Exec(
		`DELETE FROM `+tableNameQuoted+`
		WHERE rid = $1`,
		rid,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r UserDBHandler) SelectData(projectRid uuid.UUID, datamodelRid uuid.UUID, rid uuid.UUID) (*model.Data, error) {
	data := &model.User{}

	row := r.db.Instance.QueryRow(
		`SELECT
			id,
			rid,
			// TODO add columns
			created_at,
			updated_at
		FROM
			user
		WHERE
			rid = $1`,
		rid,
	)
	err := row.Scan(
		&data.ID,
		&data.RID,
		// TODO add User columns
		&data.CreatedAt,
		&data.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r DataDBHandler) SelectAllData(projectRid uuid.UUID, datamodelRid uuid.UUID, lastId int, entries int) ([]*model.Data, error) {
	var datas []*model.Data

	tableNameQuoted := pq.QuoteIdentifier(r.tableNameByProjectRidAndDatamodelRid(projectRid, datamodelRid))
	rows, err := r.db.Instance.Query(
		`SELECT
			id,
			rid,
			owner_rid,
			COALESCE(parent_id, 0),
			parent_rid,
			parent_column,
			details,
			created_at,
			updated_at
		FROM
			`+tableNameQuoted+`
		WHERE (0 = $1
			OR created_at < (
				SELECT
					d.created_at
				FROM
					`+tableNameQuoted+` AS d
				WHERE
					d.id = $1))
		ORDER BY
			created_at DESC
		LIMIT $2`,
		lastId,
		entries,
	)
	if err != nil {
		return []*model.Data{}, err
	}

	defer rows.Close()

	for rows.Next() {
		data := &model.Data{}
		err := rows.Scan(
			&data.ID,
			&data.RID,
			&data.OwnerRID,
			&data.ParentID,
			&data.ParentRID,
			&data.ParentColumn,
			&data.Details,
			&data.CreatedAt,
			&data.UpdatedAt,
		)
		if err != nil {
			return []*model.Data{}, err
		}

		datas = append(datas, data)
	}

	return datas, nil
}

func (r DataDBHandler) SelectAllDataBySearch(projectRid uuid.UUID, datamodelRid uuid.UUID, search string, lastId int, entries int) ([]*model.Data, error) {
	var datas []*model.Data

	log.Printf("search: %v", search)

	rows, err := r.db.Instance.Query(
		`SELECT
			id,
			rid,
			owner_rid,
			COALESCE(parent_id, 0),
			parent_rid,
			parent_column,
			details,
			created_at,
			updated_at
		FROM user 
		WHERE (key ILIKE '%' || $1 || '%'
				OR name ILIKE '%' || $1 || '%'
				OR description ILIKE '%' || $1 || '%')
			AND (0 = $2
				OR created_at < (
					SELECT
						u.created_at
					FROM
						user AS u
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
		return []*model.Data{}, err
	}

	defer rows.Close()

	for rows.Next() {
		data := &model.Data{}
		err := rows.Scan(
			&data.ID,
			&data.RID,
			&data.OwnerRID,
			&data.ParentID,
			&data.ParentRID,
			&data.ParentColumn,
			&data.Details,
			&data.CreatedAt,
			&data.UpdatedAt,
		)
		if err != nil {
			return []*model.Data{}, err
		}

		datas = append(datas, data)
	}

	return datas, err
}
