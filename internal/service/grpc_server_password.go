package service

import (
	"context"

	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/service/dbservice"
	pb "github.com/mrechkunov/goKeeper.git/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SavePassword save password in storage
func (gk *GoKeeperServer) SavePassword(ctx context.Context, in *pb.PasswordData) (*pb.EmptyMessage, error) {
	data := model.Passwords{
		UserLogin: in.GetLogin(),
		Pair:      in.GetPair(),
		MetaData:  in.GetMetadata(),
	}

	if err := dbservice.AddPassword(ctx, data); err != nil {
		logger.Log.Warnln("Error while save password in db", err)
		return nil, status.Error(codes.AlreadyExists, "server error pass not saved")
	}

	return &pb.EmptyMessage{}, nil
}

// GetPassword return password from storage
func (gk *GoKeeperServer) GetPassword(ctx context.Context, in *pb.PasswordData) (*pb.PasswordData, error) {
	data, err := dbservice.GetPassword(ctx, in.GetLogin(), in.GetMetadata())
	if err != nil {
		logger.Log.Warnln("error while get pass", err)
		return nil, err
	}
	out := pb.PasswordData_builder{
		Login:    &data.UserLogin,
		Pair:     &data.Pair,
		Metadata: &data.MetaData,
	}.Build()
	return out, nil
}

// EditPassword edit password in storage
func (gk *GoKeeperServer) EditPassword(ctx context.Context, in *pb.PasswordData) (*pb.EmptyMessage, error) {
	data := model.Passwords{
		UserLogin: in.GetLogin(),
		Pair:      in.GetPair(),
		MetaData:  in.GetMetadata(),
	}

	if err := dbservice.EditPassword(ctx, data); err != nil {
		logger.Log.Warnln("error while edit pass", err)
		return nil, err
	}

	return &pb.EmptyMessage{}, nil
}

// DeletePassword delete password from storage
func (gk *GoKeeperServer) DeletePassword(ctx context.Context, in *pb.PasswordData) (*pb.EmptyMessage, error) {
	data := model.Passwords{
		UserLogin: in.GetLogin(),
		Pair:      in.GetPair(),
		MetaData:  in.GetMetadata(),
	}

	if err := dbservice.DeletePassword(ctx, data); err != nil {
		logger.Log.Warnln("error while delete pass", err)
		return nil, err
	}

	return &pb.EmptyMessage{}, nil
}
