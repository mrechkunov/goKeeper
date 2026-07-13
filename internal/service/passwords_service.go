package service

// TODO: read record (uuid + metadata)
// TODO: insert recors (model.password)
// TODO: edit record (model.password)
// TODO: delete record (uuid + metadata)
// TODO: delete records (uuid)

// func GetUserByLogin(ctx context.Context, login string) model.Users {
// 	storageUsers := repository.NewUsersStorage(config.DBconn)
// 	return storageUsers.GetUserByLogin(ctx, uuid)
// }

// func GetUserByLogin(ctx context.Context, login string) model.Users {
// 	storageUsers := repository.NewUsersStorage(config.DBconn)
// 	return storageUsers.GetUserByLogin(ctx, login)
// }

// func UpdateUser(ctx context.Context, user *model.Users) error {
// 	storageUsers := repository.NewUsersStorage(config.DBconn)
// 	return storageUsers.UpdateUser(ctx, *user)
// }

// func InsertNewUser(ctx context.Context, user *model.Users) error {
// 	storageUsers := repository.NewUsersStorage(config.DBconn)
// 	if storageUsers.InsertUser(ctx, *user) != nil {
// 		return storageUsers.InsertUser(ctx, *user)
// 	}
// 	storageBalance := repository.NewBalanceStorage(config.DBconn)
// 	if storageBalance.AddUserBalance(ctx, user.Login) != nil {
// 		return storageBalance.AddUserBalance(ctx, user.Login)
// 	}
// 	return nil
// }
