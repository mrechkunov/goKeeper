package dbservice

import (
	"context"

	"github.com/mrechkunov/goKeeper.git/internal/config"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	"github.com/mrechkunov/goKeeper.git/internal/repository"
)

// AddCard insert data in storage
func AddCard(ctx context.Context, data model.Cards) error {
	cardStorage := repository.NewCardsStorage(config.DBconn)
	return cardStorage.InsertCard(ctx, data)
}

// GetCard return data from storage selected by login & metadata
func GetCard(ctx context.Context, login, metadata string) (data model.Cards, err error) {
	cardStorage := repository.NewCardsStorage(config.DBconn)
	return cardStorage.SelectCard(ctx, login, metadata)
}

// EditCard data in DB
func EditCard(ctx context.Context, dataIn model.Cards) error {
	cardStorage := repository.NewCardsStorage(config.DBconn)
	return cardStorage.UpdateCard(ctx, dataIn)
}

// DeleteCard delete row with card data by login and metadata
func DeleteCard(ctx context.Context, data model.Cards) error {
	cardStorage := repository.NewCardsStorage(config.DBconn)
	return cardStorage.DeleteCard(ctx, data)
}

// DeleteAllUserCards delete all records by login
func DeleteAllUserCards(ctx context.Context, login string) error {
	cardStorage := repository.NewCardsStorage(config.DBconn)
	return cardStorage.DeleteAllCardsByLogin(ctx, login)
}
