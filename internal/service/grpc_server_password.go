package service

import (
	"context"

	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/service/db"
	pb "github.com/mrechkunov/goKeeper.git/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SavePassword save password in storage
func (gk *GoKeeperServer) SavePassword(ctx context.Context, in *pb.PasswordData) (out *pb.EmptyMessage, err error) {
	data := model.Passwords{
		UserLogin: in.GetLogin(),
		Pair:      in.GetPair(),
		MetaData:  in.GetMetadata(),
	}
	if err = db.AddPassword(ctx, data); err != nil {
		logger.Log.Warnln("Error while save password in db", err)
		return out, status.Error(codes.AlreadyExists, "server error pass not saved")
	}
	return out, nil
}

// GetPassword return password from storage
func (gk *GoKeeperServer) GetPassword(ctx context.Context, in *pb.PasswordData) (out *pb.PasswordData, err error) {
	data, err := db.GetPassword(ctx, in.GetLogin(), in.GetMetadata())
	if err != nil {
		logger.Log.Warnln("error while get pass", err)
	}
	out = pb.PasswordData_builder{
		Login:    &data.UserLogin,
		Pair:     &data.Pair,
		Metadata: &data.MetaData,
	}.Build()
	return out, nil
}

// EditPassword edit password in storage
func (gk *GoKeeperServer) EditPassword(ctx context.Context, in *pb.PasswordData) (out *pb.EmptyMessage, err error) {
	data := model.Passwords{
		UserLogin: in.GetLogin(),
		Pair:      in.GetPair(),
		MetaData:  in.GetMetadata(),
	}
	err = db.EditPassword(ctx, data)
	if err != nil {
		logger.Log.Warnln("error while edit pass", err)
		return out, err
	}
	return out, nil
}

// DeletePassword delete password from storage
func (gk *GoKeeperServer) DeletePassword(ctx context.Context, in *pb.PasswordData) (out *pb.EmptyMessage, err error) {
	data := model.Passwords{
		UserLogin: in.GetLogin(),
		Pair:      in.GetPair(),
		MetaData:  in.GetMetadata(),
	}
	err = db.DeletePassword(ctx, data)
	if err != nil {
		logger.Log.Warnln("error while delete pass", err)
		return out, err
	}
	return out, nil
}
