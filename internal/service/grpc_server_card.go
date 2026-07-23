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

// SaveCard save card data in storage
func (gk *GoKeeperServer) SaveCard(ctx context.Context, in *pb.CardData) (*pb.EmptyMessage, error) {
	data := model.Cards{
		UserLogin:  in.GetLogin(),
		CipherData: in.GetCipherdata(),
		MetaData:   in.GetMetadata(),
	}

	if err := dbservice.AddCard(ctx, data); err != nil {
		logger.Log.Warnln("Error while save card in db", err)
		return nil, status.Error(codes.AlreadyExists, "server error card not saved")
	}

	return &pb.EmptyMessage{}, nil
}

// GetCard return card data from storage
func (gk *GoKeeperServer) GetCard(ctx context.Context, in *pb.CardData) (*pb.CardData, error) {
	data, err := dbservice.GetCard(ctx, in.GetLogin(), in.GetMetadata())
	if err != nil {
		logger.Log.Warnln("error while get card data", err)
		return nil, err
	}

	out := pb.CardData_builder{
		Login:      &data.UserLogin,
		Cipherdata: &data.CipherData,
		Metadata:   &data.MetaData,
	}.Build()

	return out, nil
}

// EditCard edit card data in storage
func (gk *GoKeeperServer) EditCard(ctx context.Context, in *pb.CardData) (*pb.EmptyMessage, error) {
	data := model.Cards{
		UserLogin:  in.GetLogin(),
		CipherData: in.GetCipherdata(),
		MetaData:   in.GetMetadata(),
	}

	if err := dbservice.EditCard(ctx, data); err != nil {
		logger.Log.Warnln("error while edit card data", err)
		return nil, err
	}

	return &pb.EmptyMessage{}, nil
}

// DeleteCard delete card data from storage
func (gk *GoKeeperServer) DeleteCard(ctx context.Context, in *pb.CardData) (*pb.EmptyMessage, error) {
	data := model.Cards{
		UserLogin:  in.GetLogin(),
		CipherData: in.GetCipherdata(),
		MetaData:   in.GetMetadata(),
	}

	if err := dbservice.DeleteCard(ctx, data); err != nil {
		logger.Log.Warnln("error while delete card", err)
		return nil, err
	}

	return &pb.EmptyMessage{}, nil
}
