package cliservice

import (
	"context"

	"github.com/mrechkunov/goKeeper.git/internal/cryptodata"
	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	pb "github.com/mrechkunov/goKeeper.git/proto"
)

// SaveCard client service for save card data on server
func SaveCard(ctx context.Context, client pb.GoKeeperClient, data model.Cards) (err error) {
	data.CipherData, err = cryptodata.CryptoCard(data.CardNumber, data.ValidTo, data.CVVCode)
	if err != nil {
		logger.Log.Errorln("failed to encrypt card data:", err)
		return err
	}
	dataPb := pb.CardData_builder{
		Login:      &data.UserLogin,
		Cipherdata: &data.CipherData,
		Metadata:   &data.MetaData,
	}.Build()
	_, err = client.SaveCard(ctx, dataPb)
	if err != nil {
		return err
	}
	return nil
}

// GetCard return card data from server
func GetCard(ctx context.Context, client pb.GoKeeperClient, card model.Cards) (out model.Cards, err error) {
	cardPb := pb.CardData_builder{
		Login:      &card.UserLogin,
		Cipherdata: &card.CipherData,
		Metadata:   &card.MetaData,
	}.Build()
	data, err := client.GetCard(ctx, cardPb)
	if err != nil {
		return out, err
	}
	number, valid, cvv, err := cryptodata.DecryptCard(data.GetCipherdata())
	if err != nil {
		return out, err
	}
	out = model.Cards{
		UserLogin:  data.GetLogin(),
		MetaData:   data.GetMetadata(),
		CardNumber: number,
		ValidTo:    valid,
		CVVCode:    cvv,
	}
	return out, err
}

// EditCard edit card data on server
func EditCard(ctx context.Context, client pb.GoKeeperClient, card model.Cards) (err error) {
	card.CipherData, err = cryptodata.CryptoCard(card.CardNumber, card.ValidTo, card.CVVCode)
	if err != nil {
		logger.Log.Errorln("failed to encrypt card data during edit:", err)
		return err
	}
	cardPb := pb.CardData_builder{
		Login:      &card.UserLogin,
		Cipherdata: &card.CipherData,
		Metadata:   &card.MetaData,
	}.Build()
	_, err = client.EditCard(ctx, cardPb)
	if err != nil {
		return err
	}
	return nil
}

// DeleteCard delete card data from server
func DeleteCard(ctx context.Context, client pb.GoKeeperClient, card model.Cards) (err error) {
	cardPb := pb.CardData_builder{
		Login:    &card.UserLogin,
		Metadata: &card.MetaData,
	}.Build()
	_, err = client.DeleteCard(ctx, cardPb)
	if err != nil {
		return err
	}
	return nil
}
