package service

import (
	"context"

	"github.com/mrechkunov/goKeeper.git/internal/model"
	pb "github.com/mrechkunov/goKeeper.git/proto"
)

// SavePass client service for save password and login
func SavePass(ctx context.Context, client pb.GoKeeperClient, pass model.Passwords) error {
	passPb := pb.PasswordData_builder{
		Login:    &pass.Login,
		Pair:     &pass.Pair,
		Metadata: &pass.Metadata,
	}.Build()
	_, err := client.SavePassword(ctx, passPb)
	if err != nil {
		return err
	}
	return nil
}

// edit pass
// delete pass
