package cliservice

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/mrechkunov/goKeeper.git/internal/cryptodata"
	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/model"
)

// SaveFile client service for save file data on server
func SaveFile(ctx context.Context, file model.File) (err error) {
	//узнать размер файла, если он больше чем 4 мб отказать
	fileInfo, err := os.Stat(file.FilePath)
	if err != nil {
		logger.Log.Infoln("error while file info get in os Stat:", err)
		return
	}
	if fileInfo.Size() > 4000000 {
		err = errors.New("to big file to save")
		logger.Log.Infoln(err)
		return err
	}
	file.FileName = filepath.Base(file.FilePath)
	//прочитать файл в байты
	data, err := os.ReadFile(file.FilePath)
	if err != nil {
		logger.Log.Warnln("error while reading file", err)
		return err
	}

	//зашифровать байты
	file.CipherData, err = cryptodata.CryptoFile(data)
	// передать на сервер данные

	return nil
}

// 	data.CipherData, err = cryptodata.CryptoCard(data.CardNumber, data.ValidTo, data.CVVCode)
// 	if err != nil {
// 		logger.Log.Infoln(err)
// 	}
// 	dataPb := pb.CardData_builder{
// 		Login:      &data.UserLogin,
// 		Cipherdata: &data.CipherData,
// 		Metadata:   &data.MetaData,
// 	}.Build()
// 	_, err = client.SaveCard(ctx, dataPb)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// // Get Card return card data from server
// func GetCard(ctx context.Context, client pb.GoKeeperClient, card model.Cards) (out model.Cards, err error) {
// 	cardPb := pb.CardData_builder{
// 		Login:      &card.UserLogin,
// 		Cipherdata: &card.CipherData,
// 		Metadata:   &card.MetaData,
// 	}.Build()
// 	data, err := client.GetCard(ctx, cardPb)
// 	if err != nil {
// 		return out, err
// 	}
// 	number, valid, cvv, err := cryptodata.DecryptCard(data.GetCipherdata())
// 	if err != nil {
// 		return out, err
// 	}
// 	out = model.Cards{
// 		UserLogin:  data.GetLogin(),
// 		MetaData:   data.GetMetadata(),
// 		CardNumber: number,
// 		ValidTo:    valid,
// 		CVVCode:    cvv,
// 	}
// 	return out, err
// }

// // EditCard edit card data on server
// func EditCard(ctx context.Context, client pb.GoKeeperClient, card model.Cards) (err error) {
// 	card.CipherData, err = cryptodata.CryptoCard(card.CardNumber, card.ValidTo, card.CVVCode)
// 	if err != nil {
// 		logger.Log.Infoln(err)
// 	}
// 	cardPb := pb.CardData_builder{
// 		Login:      &card.UserLogin,
// 		Cipherdata: &card.CipherData,
// 		Metadata:   &card.MetaData,
// 	}.Build()
// 	_, err = client.EditCard(ctx, cardPb)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// // DeleteCard delete card data from server
// func DeleteCard(ctx context.Context, client pb.GoKeeperClient, card model.Cards) (err error) {
// 	cardPb := pb.CardData_builder{
// 		Login:    &card.UserLogin,
// 		Metadata: &card.MetaData,
// 	}.Build()
// 	_, err = client.DeleteCard(ctx, cardPb)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
