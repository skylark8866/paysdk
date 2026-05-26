package service

import (
	"errors"
	"shop-demo/model"
	"shop-demo/repo"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo      *repo.Repository
	jwtSecret []byte
}

func NewUserService(repo *repo.Repository, jwtSecret string) *UserService {
	return &UserService{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *UserService) Register(username, password string) (*model.User, string, error) {
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 20 {
		return nil, "", errors.New("用户名长度需 3-20 个字符")
	}
	if len(password) < 6 {
		return nil, "", errors.New("密码长度至少 6 个字符")
	}

	existing, _ := s.repo.GetUserByUsername(username)
	if existing != nil {
		return nil, "", errors.New("用户名已存在")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", errors.New("密码加密失败")
	}

	user := &model.User{
		Username: username,
		Password: string(hashedPassword),
		Balance:  0,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, "", errors.New("创建用户失败")
	}

	token, err := s.generateToken(user.ID, user.Username)
	if err != nil {
		return nil, "", errors.New("生成令牌失败")
	}

	return user, token, nil
}

func (s *UserService) Login(username, password string) (*model.User, string, error) {
	user, err := s.repo.GetUserByUsername(strings.TrimSpace(username))
	if err != nil {
		return nil, "", errors.New("用户名或密码错误")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", errors.New("用户名或密码错误")
	}

	token, err := s.generateToken(user.ID, user.Username)
	if err != nil {
		return nil, "", errors.New("生成令牌失败")
	}

	return user, token, nil
}

func (s *UserService) GetUserByID(id uint64) (*model.User, error) {
	return s.repo.GetUserByID(id)
}

type Claims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (s *UserService) generateToken(userID uint64, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *UserService) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("无效令牌")
}
