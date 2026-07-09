package service

import (
	"context"

	"github.com/mrechkunov/goKeeper.git/internal/logger"
	"github.com/mrechkunov/goKeeper.git/internal/model"
	pb "github.com/mrechkunov/goKeeper.git/proto"
)

type GoKeeperServer struct {
	pb.UnimplementedGoKeeperServer
}

func (gk *GoKeeperServer) RegisterUser(ctx context.Context, in *pb.User) (out *pb.StatusResponce, err error) {
	user := model.Users{
		Login:    in.GetLogin(),
		Password: in.GetPassword(),
		Token:    in.GetToken(),
	}
	err = InsertUser(ctx, user)
	if err != nil {
		logger.Log.Infoln("Error while insert user:", err)
		return out, err
	}
	result := "OK"
	out = pb.StatusResponce_builder{
		Result: &result,
	}.Build()

	return out, nil
}

// // ListUserURLs return to user all urls where user is creator
// func (g *GoKeeperServer) ListUserURLs(ctx context.Context, e *pb.EmptyMessage) (*pb.UserURLsResponse, error) {
// 	// читаем метаданные и извлекаем токен
// 	md, ok := metadata.FromIncomingContext(ctx)
// 	if !ok {
// 		return nil, status.Error(codes.InvalidArgument, "no metadata found")
// 	}
// 	// берем uid из токена
// 	var uid uint32
// 	var err error
// 	if values := md["authorization"]; len(values) > 0 {
// 		token := values[0]
// 		uid, err = cryptoauth.GetIDFromCookie(token)
// 		if err != nil {
// 			logger.Log.Infoln("no ID in metadata")
// 			return nil, status.Error(codes.InvalidArgument, "error while get ID in metadata")
// 		}
// 	} else {
// 		logger.Log.Infoln("Unauthenticated")
// 		return nil, status.Error(codes.Unauthenticated, "Unauthenticated User")
// 	}
// 	// Выбрать из хранилища все записи с uid
// 	responseBatch := repository.GetStorage().GetDataByUID(uid)
// 	// добавляем всем префикс базового адреса
// 	baseResultAdress := config.ConfigAdreses.ResultServerAdress
// 	var result []*pb.URLData
// 	for _, rb := range responseBatch {
// 		urlData := &pb.URLData{}
// 		urlData.SetOriginalUrl(rb.OriginalURL)
// 		urlData.SetShortUrl(baseResultAdress + "/" + rb.ShortURL)
// 		result = append(result, urlData)
// 	}

// 	out := pb.UserURLsResponse_builder{
// 		Url: result,
// 	}.Build()
// 	return out, nil
// }

// // ExplandURLS return original url if short url is exist in DB and not set as deleted
// func (g *ShortenerServer) ExpandURL(ctx context.Context, in *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
// 	// читаем метаданные и извлекаем токен
// 	md, ok := metadata.FromIncomingContext(ctx)
// 	if !ok {
// 		return nil, status.Error(codes.Unauthenticated, "no metadata found")
// 	}
// 	// проверяем токен
// 	if values := md["authorization"]; len(values) > 0 {
// 		token := values[0]
// 		isExist := repository.GetStorage().IsCookieExist(token)
// 		if !isExist {
// 			logger.Log.Infoln("not authorizated user")
// 			return nil, status.Error(codes.Unauthenticated, "not authorizated user")
// 		}
// 	}
// 	shortURL := in.GetId()
// 	shortURL = shortURL[1:]
// 	longURL, isFound := repository.GetStorage().GetData(shortURL)
// 	if repository.GetStorage().IsDeleted(shortURL) {
// 		return nil, status.Error(codes.NotFound, "short URL is deleted")
// 	}
// 	if !isFound {
// 		return nil, status.Error(codes.NotFound, "short URL not found")
// 	}

// 	out := pb.URLExpandResponse_builder{
// 		Result: &longURL,
// 	}.Build()
// 	return out, nil
// }

// // ShortenURL shorting url from json and insert it in DB
// func (g *ShortenerServer) ShortenURL(ctx context.Context, in *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
// 	baseResultAdress := config.ConfigAdreses.ResultServerAdress
// 	// проверяем токен
// 	var token string
// 	// читаем метаданные и извлекаем токен
// 	md, ok := metadata.FromIncomingContext(ctx)
// 	if !ok {
// 		return nil, status.Error(codes.InvalidArgument, "no metadata found")
// 	}
// 	if values := md["authorization"]; len(values) > 0 {
// 		token = values[0]
// 		isExist := repository.GetStorage().IsCookieExist(token)
// 		if !isExist {
// 			logger.Log.Infoln("not authorizated user")
// 			return nil, status.Error(codes.InvalidArgument, "not authorizated user")
// 		}
// 	}
// 	originalUrl := in.GetUrl()
// 	//сокращаем url
// 	hash := sha256.Sum256([]byte(originalUrl))
// 	shortstr := hex.EncodeToString(hash[:4])
// 	ShortURL := baseResultAdress + "/" + shortstr // 4 байта хеша = 8 символов в hex
// 	out := pb.URLShortenResponse_builder{
// 		Result: &ShortURL,
// 	}.Build()
// 	// пишем в хранилище
// 	err := repository.GetStorage().SetData(shortstr, originalUrl, token)
// 	if err != nil {
// 		return nil, status.Error(codes.AlreadyExists, "ulr is exist")
// 	}

// 	return out, nil
// }
