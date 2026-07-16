package service

import (
	"context"
	"errors"

	"github.com/mrechkunov/goKeeper.git/internal/auth"
	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	pb "github.com/mrechkunov/goKeeper.git/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type GoKeeperServer struct {
	pb.UnimplementedGoKeeperServer
}

// RegisterUser Register new user if not exist in DB
func (gk *GoKeeperServer) RegisterUser(ctx context.Context, in *pb.User) (out *pb.StatusResponce, err error) {
	pass, err := auth.HashPassword(in.GetPasswordHash())
	if err != nil {
		return
	}
	user := model.Users{
		Login:        in.GetLogin(),
		PasswordHash: pass,
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
	// Создаем исходящие метаданные для заголовков (headers)
	headerMD := metadata.Pairs(
		"authorization", "Bearer "+token,
	)
	// Отправляем заголовки клиенту
	grpc.SetHeader(ctx, headerMD)
	return out, nil
}

// AuthenticateUser возвращает по логину и hash паролю token в md
func (gk *GoKeeperServer) AuthenticateUser(ctx context.Context, in *pb.User) (out *pb.User, err error) {
	user, err := GetUserByLogin(ctx, in.GetLogin())
	if err != nil {
		return out, err
	}
	if !auth.CheckPasswordHash(in.GetPasswordHash(), user.PasswordHash) {
		return out, status.Error(codes.Unauthenticated, "wrong pair login/password")
	}
	token, err := auth.GenerateToken(user.Login)
	if err != nil {
		logger.Log.Warnln("Error while generate token", err)
	}
	// Создаем исходящие метаданные для заголовков (headers)
	headerMD := metadata.Pairs(
		"authorization", "Bearer "+token,
	)
	// Отправляем заголовки клиенту
	grpc.SetHeader(ctx, headerMD)
	out = pb.User_builder{
		Login:        &user.Login,
		PasswordHash: &user.PasswordHash,
	}.Build()
	return out, nil
}
func (gk *GoKeeperServer) EditUser(ctx context.Context, in *pb.User) (out *pb.User, err error) {
	pass, err := auth.HashPassword(in.GetPasswordHash())
	if err != nil {
		return
	}

	user := model.Users{
		Login:        ctx.Value(userLoginKey).(string),
		PasswordHash: pass,
	}
	out = pb.User_builder{
		Login:        &user.Login,
		PasswordHash: &user.PasswordHash,
	}.Build()
	return out, EditUser(ctx, user)

}

func (gk *GoKeeperServer) GetPassHash(ctx context.Context, in *pb.User) (out *pb.User, err error) {
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

func (gk *GoKeeperServer) DeleteUser(ctx context.Context, in *pb.User) (out *pb.StatusResponce, err error) {
	user := model.Users{
		Login:        in.GetLogin(),
		PasswordHash: in.GetPasswordHash(),
	}
	if err = UserDelete(ctx, user); err != nil {
		return out, status.Error(codes.Internal, "server error user not deleted")
	}
	result := "user deleted"
	out = pb.StatusResponce_builder{
		Result: &result,
	}.Build()
	return out, nil
}
