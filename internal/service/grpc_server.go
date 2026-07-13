package service

import (
	"context"

	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	pb "github.com/mrechkunov/goKeeper.git/proto"
)

type GoKeeperServer struct {
	pb.UnimplementedGoKeeperServer
}

func (gk *GoKeeperServer) RegisterUser(ctx context.Context, in *pb.User) (out *pb.StatusResponce, err error) {
	user := model.Users{
		Login:        in.GetLogin(),
		PasswordHash: in.GetPassworHash(),
		Uuid:         in.GetUuid(),
	}
	err = InsertUser(ctx, user)
	if err != nil {
		logger.Log.Infoln("Error while insert user:", err)
		return out, err
	}
	result := "OK"
	out = pb.StatusResponce_builder{
		Result: &result,
	}.Build()

	return out, nil
}

// AuthenticateUser возвращает по логину hash пароля пользователя
func (gk *GoKeeperServer) AuthenticateUser(ctx context.Context, in *pb.User) (out *pb.User, err error) {
	user, err := GetUserByLogin(ctx, in.GetLogin())
	if err != nil {
		return out, err
	}
	out = pb.User_builder{
		Login:       &user.Login,
		PassworHash: &user.PasswordHash,
		Uuid:        &user.Uuid,
	}.Build()
	return out, nil
}

// AuthorizateUser возвращает пользователя с полем Uuid + передает Uuid в поле MD
func (gk *GoKeeperServer) AuthorizateUser(ctx context.Context, in *pb.User) (out *pb.User, err error) {
	user, err := GetUserByLogin(ctx, in.GetLogin())
	// метаданные создать
	if err != nil {
		return out, err
	}
	out = pb.User_builder{
		Login:       &user.Login,
		PassworHash: &user.PasswordHash,
		Uuid:        &user.Uuid,
	}.Build()
	return out, nil
}
