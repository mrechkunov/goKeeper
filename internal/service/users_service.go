package service

// TODO: autorization user (token check) extend

// TODO: register user
// TODO: authentification user (login + pass check, responce token)
// TODO: edit user (change password for autorizated user)
// TODO: delete user (delete user and all data in torages)

// func GetUserByLogin(ctx context.Context, login string) model.Users {
// 	storageUsers := repository.NewUsersStorage(config.DBconn)
// 	return storageUsers.GetUserByLogin(ctx, token)
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
