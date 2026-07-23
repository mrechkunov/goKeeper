package service

import (
	"context"
	"errors"

	"github.com/mrechkunov/goKeeper.git/internal/auth"
	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/service/dbservice"
	pb "github.com/mrechkunov/goKeeper.git/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type GoKeeperServer struct {
	pb.UnimplementedGoKeeperServer
}

// GetPassHash return password hash from storage
func (gk *GoKeeperServer) GetPassHash(ctx context.Context, in *pb.User) (out *pb.User, err error) {
	user, err := dbservice.GetUser(ctx, in.GetLogin())
	if err != nil {
		return out, err
	}
	out = pb.User_builder{
		Login:        &user.Login,
		PasswordHash: &user.PasswordHash,
	}.Build()
	return out, nil
}

// RegisterUser Register new user if not exist in DB
func (gk *GoKeeperServer) RegisterUser(ctx context.Context, in *pb.User) (out *pb.EmptyMessage, err error) {
	passHash, err := auth.HashPassword(in.GetPasswordHash())
	if err != nil {
		return out, err
	}
	user := model.Users{
		Login:        in.GetLogin(),
		PasswordHash: passHash,
	}
	if user.PasswordHash == "" {
		err = errors.New("empty pass")
		return out, err
	}
	err = dbservice.AddUser(ctx, user)
	if err != nil {
		logger.Log.Infoln("Error while insert user:", err)
		return out, err
	}
	// генерируем token
	token, err := auth.GenerateToken(user.Login)
	if err != nil {
		return out, err
	}
	// Создаем исходящие метаданные для заголовков (headers)
	headerMD := metadata.Pairs(
		"authorization", token,
	)
	// Отправляем заголовки клиенту
	grpc.SetHeader(ctx, headerMD)
	return out, nil
}

// AuthenticateUser возвращает по логину и hash паролю token в md
func (gk *GoKeeperServer) AuthenticateUser(ctx context.Context, in *pb.User) (out *pb.EmptyMessage, err error) {
	user, err := dbservice.GetUser(ctx, in.GetLogin())
	if err != nil {
		return out, err
	}
	if !auth.CheckPasswordHash(in.GetPasswordHash(), user.PasswordHash) {
		return out, status.Error(codes.Unauthenticated, "wrong pair login/password")
	}
	token, err := auth.GenerateToken(user.Login)
	if err != nil {
		logger.Log.Warnln("Error while generate token", err)
		return out, status.Error(codes.Internal, "internal error")
	}
	// Создаем исходящие метаданные для заголовков (headers)
	headerMD := metadata.Pairs(
		"authorization", token,
	)
	// Отправляем заголовки клиенту
	grpc.SetHeader(ctx, headerMD)
	return out, nil
}

func (gk *GoKeeperServer) EditUser(ctx context.Context, in *pb.User) (*pb.User, error) {
	pass, err := auth.HashPassword(in.GetPasswordHash())
	if err != nil {
		return nil, err
	}
	// Безопасное извлечение логина из контекста
	loginVal := ctx.Value(userLoginKey)
	loginStr, ok := loginVal.(string)
	if !ok || loginStr == "" {
		return nil, status.Error(codes.Unauthenticated, "user login not found in context")
	}
	user := model.Users{
		Login:        loginStr,
		PasswordHash: pass,
	}
	err = dbservice.EditUser(ctx, user)
	if err != nil {
		return nil, err
	}
	out := pb.User_builder{
		Login:        &user.Login,
		PasswordHash: &user.PasswordHash,
	}.Build()

	return out, nil
}

func (gk *GoKeeperServer) DeleteUser(ctx context.Context, in *pb.User) (*pb.EmptyMessage, error) {
	user := model.Users{
		Login:        in.GetLogin(),
		PasswordHash: in.GetPasswordHash(),
	}
	if err := dbservice.DeleteUser(ctx, user); err != nil {
		return &pb.EmptyMessage{}, status.Error(codes.Internal, "server error user not deleted")
	}
	return &pb.EmptyMessage{}, nil
}
