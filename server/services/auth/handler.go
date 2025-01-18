package auth

import (
	"fmt"
	"ht/helper"
	"ht/model"
	"ht/server/database"
	"log"
	"os"
	"time"

	"github.com/antonlindstrom/pgstore"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/siherrmann/validator"
)

type AuthService struct {
	logger       *log.Logger
	authDb       AuthDBHandlerFunctions
	sessionStore *pgstore.PGStore
}

func NewAuthService(sessionStore *pgstore.PGStore) *AuthService {
	logger := log.New(os.Stdout, "auth: ", log.LstdFlags)
	dbConnection := database.NewDatabase(
		"auth",
		&database.DatabaseConfiguration{
			Host:     helper.GetEnvVariable("DB_AUTH_HOST"),
			Port:     helper.GetEnvVariable("DB_AUTH_PORT"),
			Database: helper.GetEnvVariable("DB_AUTH_DATABASE"),
			Username: helper.GetEnvVariable("DB_AUTH_USERNAME"),
			Password: helper.GetEnvVariable("DB_AUTH_PASSWORD"),
			Schema:   helper.GetEnvVariable("DB_AUTH_SCHEMA"),
		},
	)
	var authDb AuthDBHandlerFunctions = newAuthDBHandler(dbConnection)

	// creates main auth table
	err := authDb.CreateTable(uuid.UUID{})
	if err != nil {
		log.Fatal(err.Error())
	}

	newAuthService := &AuthService{
		logger:       logger,
		authDb:       authDb,
		sessionStore: sessionStore,
	}

	return newAuthService
}

func (s *AuthService) updateSession(c echo.Context, auth model.Auth, authenticated bool) error {
	session, _ := s.sessionStore.Get(c.Request(), "auth")

	session.Values["authenticated"] = authenticated
	session.Values["email_verified"] = auth.EmailVerified
	session.Values["user_id"] = auth.RID.String()
	session.Values["created_at"] = time.Now().Unix()

	err := session.Save(c.Request(), c.Response().Writer)
	if err != nil {
		return fmt.Errorf("error saving session: %v", err)
	}
	return nil
}

func (s *AuthService) logoutSession(c echo.Context) error {
	session, _ := s.sessionStore.Get(c.Request(), "auth")

	session.Values["authenticated"] = false
	session.Values["email_verified"] = false
	session.Values["user_id"] = ""
	session.Values["created_at"] = time.Now().Unix()

	err := session.Save(c.Request(), c.Response().Writer)
	if err != nil {
		return fmt.Errorf("error saving session: %v", err)
	}
	return nil
}

func (h *AuthService) HandleRegisterWithEmail(c echo.Context) error {
	projectRid := helper.GetRequestContext(c.Request().Context()).ProjectRID

	request := &struct {
		Email    string `upd:"email, min3 max256 con@"`
		Password string `upd:"password, min8 max30 rex'^(.*[A-Z])+(.*)$' rex'^(.*[a-z])+(.*)$' rex'^(.*\\d)+(.*)$' rex'^(.*[\x60!@#$%^&*()_+={};/':\"|\\,.<>/?~-])+(.*)$'"`
	}{}
	err := validator.UnmapOrUnmarshalRequestValidateAndUpdate(c.Request(), request)
	if err != nil {
		return err
	}

	auth := &model.Auth{
		Email:        request.Email,
		PasswordHash: request.Password,
	}

	count, err := h.authDb.CountAuthByEmail(projectRid, auth.Email)
	if err != nil {
		return fmt.Errorf("error counting auth: %v", err)
	}
	if count != 0 {
		return fmt.Errorf("account already exists")
	}

	emailVerificationCodeHash, err := helper.CreateRandomString(6, helper.LettersAndNumbers)
	if err != nil {
		return fmt.Errorf("error creating email verification code: %v", err)
	}

	auth.EmailVerificationCodeHash = emailVerificationCodeHash

	auth, err = h.authDb.InsertAuth(projectRid, auth)
	if err != nil {
		return fmt.Errorf("error inserting auth: %v", err)
	}

	// TODO send registration verification email
	log.Printf("email verification code: %v", emailVerificationCodeHash)

	err = h.updateSession(c, *auth, false)
	if err != nil {
		return fmt.Errorf("error updating session: %v", err)
	}

	return nil
}

func (h *AuthService) HandleRequestNewEmailVerificationCode(c echo.Context) error {
	projectRid := helper.GetRequestContext(c.Request().Context()).ProjectRID
	userId := helper.GetCurrentUserRID(c.Request().Context())

	auth, err := h.authDb.SelectAuth(projectRid, userId)
	if err != nil {
		return fmt.Errorf("error selecting auth: %v", err)
	}
	if auth.EmailVerified {
		return fmt.Errorf("email already verified")
	}

	emailVerificationCodeHash, err := helper.CreateRandomString(6, helper.LettersAndNumbers)
	if err != nil {
		return fmt.Errorf("error creating email verification code: %v", err)
	}

	auth.EmailVerificationCodeHash = emailVerificationCodeHash

	_, err = h.authDb.UpdateAuth(projectRid, auth)
	if err != nil {
		return fmt.Errorf("error updating auth: %v", err)
	}

	// TODO send registration verification email
	log.Printf("email verification code: %v", emailVerificationCodeHash)

	return nil
}

func (h *AuthService) HandleVerifyEmail(c echo.Context) error {
	projectRid := helper.GetRequestContext(c.Request().Context()).ProjectRID
	userId := helper.GetCurrentUserRID(c.Request().Context())

	request := &struct {
		VerificationCode string `upd:"verification_code, min1"`
	}{}
	err := validator.UnmapOrUnmarshalRequestValidateAndUpdate(c.Request(), request)
	if err != nil {
		return err
	}

	auth, err := h.authDb.SelectAuth(projectRid, userId)
	if err != nil {
		return fmt.Errorf("error selecting auth: %v", err)
	}
	if auth.EmailVerified {
		return fmt.Errorf("email already verified")
	}

	valid := h.authDb.CheckEmailVerificationCodeValid(projectRid, userId, request.VerificationCode)
	if !valid {
		return fmt.Errorf("invalid email verification code")
	}

	auth.EmailVerificationCodeHash = ""
	auth.EmailVerified = true

	auth, err = h.authDb.UpdateAuth(projectRid, auth)
	if err != nil {
		return fmt.Errorf("error updating auth: %v", err)
	}

	// TODO all things to do after finshed registration

	err = h.updateSession(c, *auth, true)
	if err != nil {
		return fmt.Errorf("error updating session: %v", err)
	}

	return nil
}

func (h *AuthService) HandleLoginWithEmail(c echo.Context) error {
	projectRid := helper.GetRequestContext(c.Request().Context()).ProjectRID

	request := &struct {
		Email    string `upd:"email, min3 max256 con@"`
		Password string `upd:"password, min1"`
	}{}
	err := validator.UnmapOrUnmarshalRequestValidateAndUpdate(c.Request(), request)
	if err != nil {
		return err
	}

	auth, err := h.authDb.SelectAuthByEmailAndPassword(projectRid, request.Email, request.Password)
	if err != nil {
		return fmt.Errorf("invalid email or password")
	}

	err = h.updateSession(c, *auth, true)
	if err != nil {
		return fmt.Errorf("error updating session: %v", err)
	}

	return nil
}

func (h *AuthService) HandleRequestPasswordReset(c echo.Context) error {
	projectRid := helper.GetRequestContext(c.Request().Context()).ProjectRID

	request := &struct {
		Email string `upd:"email, min3 max256 con@"`
	}{}
	err := validator.UnmapOrUnmarshalRequestValidateAndUpdate(c.Request(), request)
	if err != nil {
		return err
	}

	auth, err := h.authDb.SelectAuthByEmail(projectRid, request.Email)
	if err != nil {
		return fmt.Errorf("error selecting auth: %v", err)
	}

	auth.PasswordResetRequestDate = time.Now()
	auth.PasswordResetCodeHash, err = helper.CreateRandomString(6, helper.OnlyNumbers)
	if err != nil {
		return fmt.Errorf("error creating password reset code: %v", err)
	}

	auth, err = h.authDb.UpdateAuth(projectRid, auth)
	if err != nil {
		return fmt.Errorf("error updating auth: %v", err)
	}

	// TODO send password reset email
	log.Printf("password reset code: %v", auth.PasswordResetCodeHash)

	err = h.updateSession(c, *auth, false)
	if err != nil {
		return fmt.Errorf("error updating session: %v", err)
	}

	return nil
}

func (h *AuthService) HandleResetPassword(c echo.Context) error {
	projectRid := helper.GetRequestContext(c.Request().Context()).ProjectRID
	userId := helper.GetCurrentUserRID(c.Request().Context())

	request := &struct {
		NewPassword          string `upd:"new_password, min8 max30 rex'^(.*[A-Z])+(.*)$' rex'^(.*[a-z])+(.*)$' rex'^(.*\\d)+(.*)$' rex'^(.*[\x60!@#$%^&*()_+={};/':\"|\\,.<>/?~-])+(.*)$'"`
		NewPasswordConfirmed string `upd:"new_password_confirmed, min8 max30 rex'^(.*[A-Z])+(.*)$' rex'^(.*[a-z])+(.*)$' rex'^(.*\\d)+(.*)$' rex'^(.*[\x60!@#$%^&*()_+={};/':\"|\\,.<>/?~-])+(.*)$'"`
		VerificationCode     string `upd:"verification_code, min1"`
	}{}
	err := validator.UnmapOrUnmarshalRequestValidateAndUpdate(c.Request(), request)
	if err != nil {
		return err
	}

	if request.NewPassword != request.NewPasswordConfirmed {
		return fmt.Errorf("passwords do not match")
	}

	auth, err := h.authDb.SelectAuth(projectRid, userId)
	if err != nil {
		return fmt.Errorf("error selecting auth: %v", err)
	}

	valid := h.authDb.CheckPasswordResetCodeValid(projectRid, auth.RID, request.VerificationCode)
	if !valid {
		return fmt.Errorf("invalid password reset code")
	}

	auth.PasswordHash = request.NewPassword

	// If a user initially registers with email, then does not verifiy his email but logs in with token he gets set to verified.
	// If a user gets created with a temporary password and logs in with a token his temporary password gets deleted.
	if !auth.EmailVerified || len(auth.PasswordTemp) != 0 {
		auth.EmailVerified = true
		auth.PasswordTemp = ""
		auth.PasswordTempRequestDate = time.Time{}

		// TODO all things to do after finshed registration
	}

	_, err = h.authDb.UpdateAuth(projectRid, auth)
	if err != nil {
		return fmt.Errorf("error updating auth: %v", err)
	}

	return nil
}

func (h *AuthService) HandleLogout(c echo.Context) error {
	err := h.logoutSession(c)
	if err != nil {
		return fmt.Errorf("error updating session: %v", err)
	}

	return nil
}

func (h *AuthService) HandleDeleteAuth(c echo.Context) error {
	// TODO check access
	h.logger.Println("deleting auth definition")

	projectRid := helper.GetRequestContext(c.Request().Context()).ProjectRID
	userRid := helper.GetCurrentUserRID(c.Request().Context())

	err := h.authDb.DeleteAuth(projectRid, userRid)
	if err != nil {
		return err
	}

	return nil
}

func (h *AuthService) HandleGetAuth(c echo.Context) (*model.Auth, error) {
	// TODO check access
	h.logger.Println("getting auth definition")

	projectRid := helper.GetRequestContext(c.Request().Context()).ProjectRID
	userRid := helper.GetCurrentUserRID(c.Request().Context())

	auth, err := h.authDb.SelectAuth(projectRid, userRid)
	if err != nil {
		return nil, err
	}

	return auth, nil
}
