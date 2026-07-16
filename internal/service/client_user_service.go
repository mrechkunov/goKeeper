package service

import (
	"context"

	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	pb "github.com/mrechkunov/goKeeper.git/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// RegisterUser client service for register new user, return token
func RegisterUser(ctx context.Context, client pb.GoKeeperClient, user model.Users) (token string, err error) {
	userPb := pb.User_builder{
		Login:        &user.Login,
		PasswordHash: &user.PasswordHash,
	}.Build()
	var header metadata.MD
	_, err = client.RegisterUser(ctx, userPb, grpc.Header(&header))
	if err != nil {
		logger.Log.Errorln("error while register user: ", err)
		return "", err
	}
	if vals := header.Get("authorization"); len(vals) > 0 {
		token = vals[0]
	}
	return token, nil
}

// AuthenticateUser client service for authenticate user, return token
func AuthenticateUser(ctx context.Context, client pb.GoKeeperClient, user model.Users) (token string, err error) {
	userPb := pb.User_builder{
		Login:        &user.Login,
		PasswordHash: &user.PasswordHash,
	}.Build()
	var header metadata.MD
	_, err = client.AuthenticateUser(ctx, userPb, grpc.Header(&header))
	if err != nil {
		logger.Log.Errorln("error while Authenticate user: ", err)
		return token, err
	}
	if vals := header.Get("authorization"); len(vals) > 0 {
		token = vals[0]
	}
	return token, err
}

// ChangePass client service for change pass to authentificated user
func ChangePass(ctx context.Context, client pb.GoKeeperClient, user model.Users) error {
	userPb := pb.User_builder{
		Login:        &user.Login,
		PasswordHash: &user.PasswordHash,
	}.Build()
	_, err := client.EditUser(ctx, userPb)
	if err != nil {
		logger.Log.Errorln("error while edit user: ", err)
		return err
	}
	return nil
}

// DeleteUser client service for delete authentificated user
func DeleteUser(ctx context.Context, client pb.GoKeeperClient, user model.Users) error {
	userPb := pb.User_builder{
		Login:        &user.Login,
		PasswordHash: &user.PasswordHash,
	}.Build()
	_, err := client.DeleteUser(ctx, userPb)
	if err != nil {
		logger.Log.Errorln("error while delete user: ", err)
		return err
	}
	return nil
}
