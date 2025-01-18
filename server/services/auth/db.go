package auth

import (
	"context"
	"fmt"
	"ht/model"
	"ht/server/database"
	"log"
	"time"

	"github.com/google/uuid"
)

type AuthDBHandlerFunctions interface {
	CreateTable() error
	DropTable() error
	CountAuthByEmail(email string) (int, error)
	CheckEmailVerificationCodeValid(rid uuid.UUID, code string) bool
	CheckPasswordResetCodeValid(rid uuid.UUID, code string) bool
	InsertAuth(auth *model.Auth) (*model.Auth, error)
	UpdateAuth(auth *model.Auth) (*model.Auth, error)
	DeleteAuth(rid uuid.UUID) error
	SelectAuth(rid uuid.UUID) (*model.Auth, error)
	SelectAuthByEmail(email string) (*model.Auth, error)
	SelectAuthByEmailAndPassword(email string, password string) (*model.Auth, error)
	SelectAllAuth(lastId int, entries int) ([]*model.Auth, error)
	SelectAllAuthBySearch(search string, lastId int, entries int) ([]*model.Auth, error)
}

type AuthDBHandler struct {
	db *database.Database
}

func newAuthDBHandler(dbConnection *database.Database) *AuthDBHandler {
	return &AuthDBHandler{
		db: dbConnection,
	}
}

func (r AuthDBHandler) CreateTable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.db.Instance.ExecContext(
		ctx,
		`CREATE EXTENSION IF NOT EXISTS pgcrypto;

		CREATE TABLE IF NOT EXISTS auth (
			id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			rid UUID UNIQUE DEFAULT gen_random_uuid(),
			email VARCHAR(254) NOT NULL,
			password_temp TEXT DEFAULT '',
			password_temp_request_date TIMESTAMP DEFAULT '2000-01-01T01:23:45Z',
			password_hash TEXT NOT NULL,
			password_reset_code_hash TEXT DEFAULT '',
			password_reset_request_date TIMESTAMP DEFAULT '2000-01-01T01:23:45Z',
			email_verification_code_hash TEXT DEFAULT '',
			email_verification_request_date TIMESTAMP DEFAULT '2000-01-01T01:23:45Z',
			email_verified boolean DEFAULT FALSE,
			email_to_change_to TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	)
	if err != nil {
		return fmt.Errorf("error creating auth table: %#v", err)
	}

	err = r.db.CreateIndex("auth", "rid")
	if err != nil {
		return err
	}

	r.db.Logger.Println("created table auth")
	return nil
}

func (r AuthDBHandler) DropTable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `DROP TABLE IF EXISTS auth`
	_, err := r.db.Instance.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error dropping auth table: %#v", err)
	}

	r.db.Logger.Println("dropped table auth")
	return nil
}

func (r AuthDBHandler) CountAuthByEmail(email string) (int, error) {
	count := 0

	err := r.db.Instance.QueryRow(
		`SELECT
			COUNT(*)
		FROM
			auth
		WHERE
			email = $1`,
		email,
	).Scan(&count)

	return count, err
}

func (r AuthDBHandler) CheckEmailVerificationCodeValid(rid uuid.UUID, code string) bool {
	exists := false

	err := r.db.Instance.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM auth WHERE rid = $1 AND email_verification_code_hash = crypt($2, email_verification_code_hash));`,
		rid,
		code,
	).Scan(&exists)
	if err != nil {
		r.db.Logger.Println(err)
		return false
	}

	return exists
}

func (r AuthDBHandler) CheckPasswordResetCodeValid(rid uuid.UUID, code string) bool {
	exists := false

	err := r.db.Instance.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM auth WHERE rid = $1 AND password_reset_code_hash = crypt($2, password_reset_code_hash));`,
		rid,
		code,
	).Scan(&exists)
	if err != nil {
		return false
	}

	return exists
}

func (r AuthDBHandler) InsertAuth(auth *model.Auth) (*model.Auth, error) {
	row := r.db.Instance.QueryRow(
		`INSERT INTO auth (email,
			password_temp,
			password_temp_request_date,
			password_hash,
			email_verification_code_hash,
			email_verification_request_date,
			email_verified)
		VALUES (lower($1),
			$2,
			$3,
			crypt($4, gen_salt('bf', 6)),
			crypt($5, gen_salt('bf', 6)),
			$6,
			$7)
		RETURNING
			rid,
			email,
			password_temp,
			password_temp_request_date,
			password_hash,
			email_verification_code_hash,
			email_verification_request_date,
			email_verified,
			created_at,
			updated_at`,
		auth.Email,
		auth.PasswordTemp,
		auth.PasswordTempRequestDate,
		auth.PasswordHash,
		auth.EmailVerificationCodeHash,
		auth.EmailVerificationRequestDate,
		auth.EmailVerified,
	)

	err := row.Scan(
		&auth.RID,
		&auth.Email,
		&auth.PasswordTemp,
		&auth.PasswordTempRequestDate,
		&auth.PasswordHash,
		&auth.EmailVerificationCodeHash,
		&auth.EmailVerificationRequestDate,
		&auth.EmailVerified,
		&auth.CreatedAt,
		&auth.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	log.Println(auth.Email, auth.PasswordHash)

	return auth, nil
}

func (r AuthDBHandler) UpdateAuth(auth *model.Auth) (*model.Auth, error) {
	row := r.db.Instance.QueryRow(
		`UPDATE
			auth
		SET
			email = lower($1),
			password_temp = $2,
			password_temp_request_date = $3,
			password_hash = CASE 
						WHEN password_hash <> $4 THEN crypt($4, gen_salt('bf', 6)) 
						ELSE password_hash 
					END,
			password_reset_code_hash = CASE 
						WHEN password_reset_code_hash <> $5 THEN crypt($4, gen_salt('bf', 6)) 
						ELSE password_reset_code_hash 
					END,
			password_reset_request_date = $6,
			email_verification_code_hash = CASE 
						WHEN email_verification_code_hash <> $7 THEN crypt($4, gen_salt('bf', 6)) 
						ELSE email_verification_code_hash 
					END,
			email_verification_request_date = $8,
			email_verified = $9,
			email_to_change_to = $10,
			updated_at = CURRENT_TIMESTAMP
		WHERE
			rid = $11
		RETURNING
			rid,
			email,
			password_temp,
			password_temp_request_date,
			password_hash,
			email_verification_code_hash,
			email_verification_request_date,
			email_verified,
			created_at,
			updated_at`,
		auth.Email,
		auth.PasswordTemp,
		auth.PasswordTempRequestDate,
		auth.PasswordHash,
		auth.PasswordResetCodeHash,
		auth.PasswordResetRequestDate,
		auth.EmailVerificationCodeHash,
		auth.EmailVerificationRequestDate,
		auth.EmailVerified,
		auth.EmailToChangeTo,
		auth.RID,
	)

	err := row.Scan(
		&auth.RID,
		&auth.Email,
		&auth.PasswordTemp,
		&auth.PasswordTempRequestDate,
		&auth.PasswordHash,
		&auth.EmailVerificationCodeHash,
		&auth.EmailVerificationRequestDate,
		&auth.EmailVerified,
		&auth.CreatedAt,
		&auth.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return auth, nil
}

func (r AuthDBHandler) DeleteAuth(rid uuid.UUID) error {
	_, err := r.db.Instance.Exec(
		`DELETE FROM auth
		WHERE rid = $1`,
		rid,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r AuthDBHandler) SelectAuth(rid uuid.UUID) (*model.Auth, error) {
	auth := &model.Auth{}

	row := r.db.Instance.QueryRow(
		`SELECT
			id,
			rid,
			email,
			password_temp,
			password_temp_request_date,
			password_hash,
			password_reset_code_hash,
			password_reset_request_date,
			email_verification_code_hash,
			email_verification_request_date,
			email_verified,
			email_to_change_to,
			created_at,
			updated_at
		FROM
			auth
		WHERE
			rid = $1`,
		rid,
	)
	err := row.Scan(
		&auth.ID,
		&auth.RID,
		&auth.Email,
		&auth.PasswordTemp,
		&auth.PasswordTempRequestDate,
		&auth.PasswordHash,
		&auth.PasswordResetCodeHash,
		&auth.PasswordResetRequestDate,
		&auth.EmailVerificationCodeHash,
		&auth.EmailVerificationRequestDate,
		&auth.EmailVerified,
		&auth.EmailToChangeTo,
		&auth.CreatedAt,
		&auth.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return auth, nil
}

func (r AuthDBHandler) SelectAuthByEmail(email string) (*model.Auth, error) {
	auth := &model.Auth{}

	row := r.db.Instance.QueryRow(
		`SELECT
			id,
			rid,
			email,
			password_temp,
			password_temp_request_date,
			password_hash,
			password_reset_code_hash,
			password_reset_request_date,
			email_verification_code_hash,
			email_verification_request_date,
			email_verified,
			email_to_change_to,
			created_at,
			updated_at
		FROM
			auth
		WHERE
			email = lower($1)`,
		email,
	)
	err := row.Scan(
		&auth.ID,
		&auth.RID,
		&auth.Email,
		&auth.PasswordTemp,
		&auth.PasswordTempRequestDate,
		&auth.PasswordHash,
		&auth.PasswordResetCodeHash,
		&auth.PasswordResetRequestDate,
		&auth.EmailVerificationCodeHash,
		&auth.EmailVerificationRequestDate,
		&auth.EmailVerified,
		&auth.EmailToChangeTo,
		&auth.CreatedAt,
		&auth.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return auth, nil
}

func (r AuthDBHandler) SelectAuthByEmailAndPassword(email string, password string) (*model.Auth, error) {
	auth := &model.Auth{}

	row := r.db.Instance.QueryRow(
		`SELECT
			id,
			rid,
			email,
			password_temp,
			password_temp_request_date,
			password_hash,
			password_reset_code_hash,
			password_reset_request_date,
			email_verification_code_hash,
			email_verification_request_date,
			email_verified,
			email_to_change_to,
			created_at,
			updated_at
		FROM
			auth
		WHERE
			email = lower($1)
		AND (password_hash = crypt($2, password_hash)
			OR (length(password_temp) > 0 AND password_temp = $2))`,
		email,
		password,
	)
	err := row.Scan(
		&auth.ID,
		&auth.RID,
		&auth.Email,
		&auth.PasswordTemp,
		&auth.PasswordTempRequestDate,
		&auth.PasswordHash,
		&auth.PasswordResetCodeHash,
		&auth.PasswordResetRequestDate,
		&auth.EmailVerificationCodeHash,
		&auth.EmailVerificationRequestDate,
		&auth.EmailVerified,
		&auth.EmailToChangeTo,
		&auth.CreatedAt,
		&auth.UpdatedAt,
	)
	if err != nil {
		log.Printf("error getting auth: %v", err)
		return nil, err
	}

	return auth, nil
}

func (r AuthDBHandler) SelectAllAuth(lastId int, entries int) ([]*model.Auth, error) {
	var auths []*model.Auth

	rows, err := r.db.Instance.Query(`
		SELECT
			id,
			rid,
			email,
			password_temp,
			password_temp_request_date,
			password_hash,
			password_reset_code_hash,
			password_reset_request_date,
			email_verification_code_hash,
			email_verification_request_date,
			email_verified,
			email_to_change_to,
			created_at,
			updated_at
		FROM
			auth
		WHERE (0 = $1
			OR created_at < (
				SELECT
					created_at
				FROM
					auth
				WHERE
					id = $1))
		ORDER BY
			created_at DESC
		LIMIT $2`,
		lastId,
		entries,
	)
	if err != nil {
		return []*model.Auth{}, err
	}

	defer rows.Close()

	for rows.Next() {
		auth := &model.Auth{}
		err := rows.Scan(
			&auth.ID,
			&auth.RID,
			&auth.Email,
			&auth.PasswordTemp,
			&auth.PasswordTempRequestDate,
			&auth.PasswordHash,
			&auth.PasswordResetCodeHash,
			&auth.PasswordResetRequestDate,
			&auth.EmailVerificationCodeHash,
			&auth.EmailVerificationRequestDate,
			&auth.EmailVerified,
			&auth.EmailToChangeTo,
			&auth.CreatedAt,
			&auth.UpdatedAt,
		)
		if err != nil {
			return []*model.Auth{}, err
		}

		auths = append(auths, auth)
	}

	return auths, nil
}

func (r AuthDBHandler) SelectAllAuthBySearch(search string, lastId int, entries int) ([]*model.Auth, error) {
	var auths []*model.Auth

	log.Printf("search: %v", search)

	rows, err := r.db.Instance.Query(`
		SELECT
			id,
			rid,
			email,
			password_temp,
			password_temp_request_date,
			password_hash,
			password_reset_code_hash,
			password_reset_request_date,
			email_verification_code_hash,
			email_verification_request_date,
			email_verified,
			email_to_change_to,
			created_at,
			updated_at
		FROM auth
		WHERE (rid ILIKE '%' || $1 || '%'
				OR email ILIKE '%' || lower($1) || '%')
			AND (0 = $2
				OR created_at < (
					SELECT
						u.created_at
					FROM
						auth AS u
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
		return []*model.Auth{}, err
	}

	defer rows.Close()

	for rows.Next() {
		auth := &model.Auth{}
		err := rows.Scan(
			&auth.ID,
			&auth.RID,
			&auth.Email,
			&auth.PasswordTemp,
			&auth.PasswordTempRequestDate,
			&auth.PasswordHash,
			&auth.PasswordResetCodeHash,
			&auth.PasswordResetRequestDate,
			&auth.EmailVerificationCodeHash,
			&auth.EmailVerificationRequestDate,
			&auth.EmailVerified,
			&auth.EmailToChangeTo,
			&auth.CreatedAt,
			&auth.UpdatedAt,
		)
		if err != nil {
			return []*model.Auth{}, err
		}

		auths = append(auths, auth)
	}

	return auths, err
}
