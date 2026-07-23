package cliservice

import (
	"context"

	"github.com/mrechkunov/goKeeper.git/internal/cryptodata"
	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	pb "github.com/mrechkunov/goKeeper.git/proto"
)

// SavePass client service for save password and login
func SavePass(ctx context.Context, client pb.GoKeeperClient, pass model.Passwords) (err error) {
	pass.Pair, err = cryptodata.CryptoPair(pass.LoginToSave, pass.PasswordToSave)
	if err != nil {
		logger.Log.Errorln("failed to encrypt password pair:", err)
		return err
	}
	passPb := pb.PasswordData_builder{
		Login:    &pass.UserLogin,
		Pair:     &pass.Pair,
		Metadata: &pass.MetaData,
	}.Build()
	_, err = client.SavePassword(ctx, passPb)
	if err != nil {
		return err
	}
	return nil
}

// Get Pass return Pair from DB
func GetPass(ctx context.Context, client pb.GoKeeperClient, pass model.Passwords) (out model.Passwords, err error) {
	passPb := pb.PasswordData_builder{
		Login:    &pass.UserLogin,
		Pair:     &pass.Pair,
		Metadata: &pass.MetaData,
	}.Build()
	data, err := client.GetPassword(ctx, passPb)
	if err != nil {
		return out, err
	}
	getLogin, getPass, err := cryptodata.DecryptPair(data.GetPair())
	if err != nil {
		return out, err
	}
	out = model.Passwords{
		UserLogin:      data.GetLogin(),
		MetaData:       data.GetMetadata(),
		LoginToSave:    getLogin,
		PasswordToSave: getPass,
	}
	return out, nil
}

// EditPass edit password pair in DB
func EditPass(ctx context.Context, client pb.GoKeeperClient, pass model.Passwords) (err error) {
	pass.Pair, err = cryptodata.CryptoPair(pass.LoginToSave, pass.PasswordToSave)
	if err != nil {
		logger.Log.Errorln("failed to encrypt password pair during edit:", err)
		return err
	}
	passPb := pb.PasswordData_builder{
		Login:    &pass.UserLogin,
		Pair:     &pass.Pair,
		Metadata: &pass.MetaData,
	}.Build()
	_, err = client.EditPassword(ctx, passPb)
	if err != nil {
		return err
	}
	return nil
}

// Delete Pass delete password pair in DB
func DeletePass(ctx context.Context, client pb.GoKeeperClient, pass model.Passwords) (err error) {
	passPb := pb.PasswordData_builder{
		Login:    &pass.UserLogin,
		Metadata: &pass.MetaData,
	}.Build()
	_, err = client.DeletePassword(ctx, passPb)
	if err != nil {
		return err
	}
	return nil
}
