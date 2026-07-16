package service

import (
	"context"

	"github.com/mrechkunov/goKeeper.git/internal/model"
	pb "github.com/mrechkunov/goKeeper.git/proto"
)

// SavePass client service for save password and login
func SavePass(ctx context.Context, client pb.GoKeeperClient, pass model.Passwords) (err error) {
	passPb := pb.PasswordData_builder{
		Login:    &pass.Login,
		Pair:     &pass.Pair,
		Metadata: &pass.Metadata,
	}.Build()
	_, err = client.SavePassword(ctx, passPb)
	if err != nil {
		return err
	}
	return nil
}

func GetPass(ctx context.Context, client pb.GoKeeperClient, pass model.Passwords) (out model.Passwords, err error) {
	passPb := pb.PasswordData_builder{
		Login:    &pass.Login,
		Pair:     &pass.Pair,
		Metadata: &pass.Metadata,
	}.Build()
	data, err := client.GetPass(ctx, passPb)
	if err != nil {
		return out, err
	}
	out = model.Passwords{
		Login:    data.GetLogin(),
		Pair:     data.GetPair(),
		Metadata: data.GetMetadata(),
	}
	return out, err
}

// var userPb pb.User
// userPb.SetLogin(user.Login)
// userPb.SetPasswordHash(user.PasswordHash)
// var header metadata.MD
// _, err = client.RegisterUser(ctx, &userPb, grpc.Header(&header))
// if err != nil {
// 	logger.Log.Errorln("error while register user: ", err)
// 	return "", err
// }
// if vals := header.Get("authorization"); len(vals) > 0 {
// 	token = vals[0]
// }
// return token, nil
