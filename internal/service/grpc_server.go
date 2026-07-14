package service

import (
	"context"
	"errors"

	"github.com/mrechkunov/goKeeper.git/internal/auth"
	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	pb "github.com/mrechkunov/goKeeper.git/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type GoKeeperServer struct {
	pb.UnimplementedGoKeeperServer
}

// RegisterUser Register new user if not exist in DB
func (gk *GoKeeperServer) RegisterUser(ctx context.Context, in *pb.User) (out *pb.StatusResponce, err error) {
	user := model.Users{
		Login:        in.GetLogin(),
		PasswordHash: in.GetPasswordHash(),
	}
	if user.PasswordHash == "" {
		err = errors.New("empty pass")
		return out, err
	}
	err = InsertUser(ctx, user)
	if err != nil {
		logger.Log.Infoln("Error while insert user:", err)
		return out, err
	}
	result := "New user sucsessful registered"
	out = pb.StatusResponce_builder{
		Result: &result,
	}.Build()
	// генерируем token
	token, err := auth.GenerateToken(user.Login)
	if err != nil {
		return
	}
	ulogin, err := auth.GetLoginByToken(token)
	if err != nil {
		logger.Log.Warnln(err)
		return
	}
	// Создаем исходящие метаданные для заголовков (headers)
	headerMD := metadata.Pairs(
		"authorization", token,
		"userlogin", ulogin,
	)
	// Отправляем заголовки клиенту
	grpc.SetHeader(ctx, headerMD)
	return out, nil
}

// AuthenticateUser возвращает по логину hash пароля пользователя
func (gk *GoKeeperServer) AuthenticateUser(ctx context.Context, in *pb.User) (out *pb.User, err error) {
	user, err := GetUserByLogin(ctx, in.GetLogin())
	if err != nil {
		return out, err
	}
	out = pb.User_builder{
		Login:        &user.Login,
		PasswordHash: &user.PasswordHash,
	}.Build()
	return out, nil
}

// AuthorizateUser возвращает пользователя с полем login + передает token в поле MD
func (gk *GoKeeperServer) AuthorizateUser(ctx context.Context, in *pb.User) (out *pb.User, err error) {
	user, err := GetUserByLogin(ctx, in.GetLogin())
	// метаданные создать
	if err != nil {
		return out, err
	}
	out = pb.User_builder{
		Login:        &user.Login,
		PasswordHash: &user.PasswordHash,
	}.Build()
	return out, nil
}
