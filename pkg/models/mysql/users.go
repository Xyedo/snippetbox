package mysql

import (
	"database/sql"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/xyedo/snippetbox/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	DB *sql.DB
}

func (u *UserModel) Insert(name, email, password string) error {
	stmt := `INSERT INTO users (name,email, hashed_password, created)
	VALUES (
		?,
		?,
		?,
		UTC_TIMESTAMP()
	)`
	ecryptedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = u.DB.Exec(stmt, name, email, string(ecryptedPass))
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "user_uc_email") {
				return models.ErrDuplicateEmail
			}
		}
		return err
	}
	return nil
}

func (u *UserModel) Authenticate(email, password string) (int, error) {
	stmt := `SELECT id, hashed_password from users WHERE email = ?`
	res := u.DB.QueryRow(stmt, email)
	var id int
	var hashedPassword []byte
	err := res.Scan(&id, &hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, models.ErrInvalidCredentials
		}
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return 0, models.ErrInvalidCredentials
		}
		return 0, err
	}

	return id, nil
}
func (u *UserModel) Get(id int) (*models.User, error) {
	return nil, nil
}
