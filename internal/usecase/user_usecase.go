package usecase

import (
	"context"
	"errors"
	"os"
	"strconv"
	"time"

	"money_plan/internal/model"
	"money_plan/internal/repository"

	jwt "github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase interface {
	Register(ctx context.Context, user *model.User) (*model.User, error)
	Login(ctx context.Context, email, password string) (string, *model.User, error)
}

type userUsecase struct {
	userRepo       repository.UserRepository
	contextTimeout time.Duration
}

func NewUserUsecase(repo repository.UserRepository) UserUsecase {
	return &userUsecase{
		userRepo:       repo,
		contextTimeout: time.Second * 5,
	}
}

func (u *userUsecase) Register(ctx context.Context, user *model.User) (*model.User, error) {
	// basic validation
	if user.Email == "" || user.Password == "" || user.Name == "" {
		return nil, errors.New("name, email and password are required")
	}

	// check existing
	if existing, err := u.userRepo.FindByEmail(ctx, user.Email); err == nil && existing.ID != 0 {
		return nil, errors.New("email already registered")
	}

	// hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.Password = string(hashed)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if err := u.userRepo.Save(ctx, user); err != nil {
		return nil, err
	}

	// hide password in response
	user.Password = ""
	return user, nil
}

func (u *userUsecase) Login(ctx context.Context, email, password string) (string, *model.User, error) {
	if email == "" || password == "" {
		return "", nil, errors.New("email and password are required")
	}

	// cari user
	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}
	if user == nil || user.ID == 0 {
		return "", nil, errors.New("invalid credentials")
	}

	// cek password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	// buat JWT
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "replace_me_in_env"
	}

	expHours := 72
	if s := os.Getenv("JWT_EXPIRATION_HOURS"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			expHours = v
		}
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(time.Duration(expHours) * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", nil, err
	}

	// hide password before returning user
	user.Password = ""
	return tokenString, user, nil
}
