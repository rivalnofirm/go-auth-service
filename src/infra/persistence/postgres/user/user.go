package user

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	dtoUser "go-auth-service/src/app/dto/user"
	"go-auth-service/src/infra/constants/common"
	errorMessage "go-auth-service/src/infra/constants/error_message"
	"go-auth-service/src/infra/helper"
	"go-auth-service/src/infra/models"
	"go-auth-service/src/infra/persistence/postgres"
	"log"
	"os"
)

type UserRepository interface {
	Create(data *dtoUser.RegisterReq) (userId int64, err error)
	GetByEmail(email string) (*models.User, error)
	GetById(id int64) (*models.User, error)
	GetUserDetailById(id int64) (*dtoUser.UserDetails, error)
	UpdateUserProfileByUserId(userId int64, firstName, lastName, birthDate, gender string) error
}

const (
	CreateUser       = `INSERT INTO user_auth (email, password) VALUES ($1, $2) RETURNING id`
	CreateUserDetail = `INSERT INTO user_detail (user_id, first_name, last_name, user_type_id) VALUES ($1, $2, $3, 3)`
	GetByEmail       = `SELECT * FROM user_auth WHERE email = $1 AND deleted_at IS NULL`
	GetById          = `SELECT * FROM user_auth WHERE id = $1 AND deleted_at IS NULL`
	GetUserDetail    = `SELECT
							ua.id,
							ua.email,
							ut.type,
							ud.first_name,
							ud.last_name,
							ud.phone,
							ud.picture,
							ud.birth_date,
							ud.gender,
							ud.verified,
							ud.created_at AS created_at,
							ud.updated_at AS updated_at,
							ud.deleted_at AS deleted_at
						FROM
							user_auth ua
						LEFT JOIN
							user_detail ud ON ua.id = ud.user_id
						LEFT JOIN
							user_type ut ON ud.user_type_id = ut.id
						WHERE
							ua.id = $1 AND ua.deleted_at IS NULL AND ud.deleted_at IS NULL`
	UpdateUserDetailByUserId = `UPDATE user_detail SET first_name = $1, last_name = $2, birth_date = $3, gender = $4, updated_at = now() WHERE user_id = $5 AND deleted_at IS NULL`
)

type PreparedStatement struct {
	createUser               *sqlx.Stmt
	createUserDetail         *sqlx.Stmt
	getByEmail               *sqlx.Stmt
	getById                  *sqlx.Stmt
	getUserDetail            *sqlx.Stmt
	updateUserDetailByUserId *sqlx.Stmt
}

type userRepo struct {
	Connection *postgres.Connection
	statement  PreparedStatement
}

func NewUserRepository(db *postgres.Connection) UserRepository {
	repo := &userRepo{
		Connection: db,
	}
	InitPreparedStatement(repo)
	return repo
}

func (p *userRepo) Preparex(query string, isMaster bool) *sqlx.Stmt {
	if !isMaster {
		// for slave
		statement, err := p.Connection.GetPrimarySlave().Preparex(query)
		if err != nil {
			log.Fatalf("Failed to preparex query: %s. Error: %s", query, err.Error())
		}

		return statement
	}

	statement, err := p.Connection.GetPrimaryMaster().Preparex(query)
	if err != nil {
		log.Fatalf("Failed to preparex query: %s. Error: %s", query, err.Error())
	}

	return statement
}

func InitPreparedStatement(m *userRepo) {
	m.statement = PreparedStatement{
		createUser:               m.Preparex(CreateUser, common.IsMasterDb),
		createUserDetail:         m.Preparex(CreateUserDetail, common.IsMasterDb),
		getByEmail:               m.Preparex(GetByEmail, common.NotIsMasterDb),
		getById:                  m.Preparex(GetById, common.NotIsMasterDb),
		getUserDetail:            m.Preparex(GetUserDetail, common.NotIsMasterDb),
		updateUserDetailByUserId: m.Preparex(UpdateUserDetailByUserId, common.IsMasterDb),
	}
}

func (p *userRepo) Create(data *dtoUser.RegisterReq) (userId int64, err error) {
	pwd, err := helper.HashPassword(data.Password)
	if err != nil {
		return 0, err
	}

	tx, err := p.Connection.GetPrimaryMaster().Beginx()
	if err != nil {
		log.Println("Failed to begin transaction:", err)
		return 0, err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Println("Recovered in Register:", r)
			err = fmt.Errorf("panic occurred: %v", r)
		} else if err != nil {
			tx.Rollback()
			log.Println("Rolling back transaction due to:", err)
		} else {
			err = tx.Commit()
			if err != nil {
				log.Println("Failed to commit transaction:", err)
			}
		}
	}()

	var resultData models.User
	err = tx.QueryRowx(CreateUser, data.Email, pwd).Scan(&resultData.Id)
	if err != nil {
		log.Println("Failed to create user:", err)
		return 0, err
	}

	_, err = tx.Exec(CreateUserDetail, resultData.Id, data.FirstName, data.LastName)
	if err != nil {
		log.Println("Failed to create user_detail:", err)
		return 0, err
	}

	return resultData.Id, nil
}

func (p *userRepo) GetByEmail(email string) (*models.User, error) {
	var user []*models.User

	err := p.statement.getByEmail.Select(&user, email)
	if err != nil {
		return nil, err
	}

	if len(user) < 1 {
		return nil, errors.New(errorMessage.UserNotFound)
	}

	return user[0], nil
}

func (p *userRepo) GetById(id int64) (*models.User, error) {
	var user []*models.User

	err := p.statement.getById.Select(&user, id)
	if err != nil {
		return nil, err
	}

	if len(user) < 1 {
		return nil, errors.New(errorMessage.UserNotFound)
	}

	return user[0], nil
}

func (p *userRepo) GetUserDetailById(id int64) (*dtoUser.UserDetails, error) {
	var user models.User
	var userDetail models.UserDetail
	var userType models.UserType

	err := p.statement.getUserDetail.QueryRow(id).Scan(
		&user.Id,
		&user.Email,
		&userType.Type,
		&userDetail.FirstName,
		&userDetail.LastName,
		&userDetail.Phone,
		&userDetail.Picture,
		&userDetail.BirthDate,
		&userDetail.Gender,
		&userDetail.Verified,
		&userDetail.CreatedAt,
		&userDetail.UpdatedAt,
		&userDetail.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New(errorMessage.UserNotFound)
		}
		return nil, err
	}

	if userDetail.Picture.String != "" {
		picturePath := fmt.Sprintf("%s%s", os.Getenv("URL_PICTURE"), userDetail.Picture.String)
		userDetail.Picture.String = picturePath
	}

	birthDate := helper.DateToStringByFormat(userDetail.BirthDate, "02-01-2006")

	createdAt := helper.DateToStringByFormat(userDetail.CreatedAt, "02-01-2006 15:04:05")
	updatedAt := helper.DateToStringByFormat(userDetail.UpdatedAt, "02-01-2006 15:04:05")
	deletedAt := helper.DateToStringByFormat(userDetail.DeletedAt, "02-01-2006 15:04:05")

	data := &dtoUser.UserDetails{
		UserId:    user.Id,
		Email:     user.Email,
		UserType:  userType.Type,
		FirstName: userDetail.FirstName.String,
		LastName:  userDetail.LastName.String,
		Phone:     userDetail.Phone.String,
		Picture:   userDetail.Picture.String,
		BirthDate: birthDate,
		Gender:    userDetail.Gender.String,
		Verified:  userDetail.Verified,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		DeletedAt: deletedAt,
	}

	return data, nil
}

func (p *userRepo) UpdateUserProfileByUserId(userId int64, firstName, lastName, birthDate, gender string) error {
	_, err := p.statement.updateUserDetailByUserId.Exec(firstName, lastName, birthDate, gender, userId)
	if err != nil {
		return err
	}

	return nil
}
