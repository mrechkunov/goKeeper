package repository

import (
	"database/sql"

	"github.com/mrechkunov/goKeeper.git/internal/config"
	"github.com/mrechkunov/goKeeper.git/internal/logger"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type DB struct {
	dbconn *sql.DB
}

// NewDB set new connection to DB from config, applying all migrations
func NewDB() *DB {
	var db DB
	var err error
	db.dbconn, err = NewConnect()
	if err != nil {
		logger.Log.Errorln("error while db connection")
	}

	m, err := migrate.New(
		config.SrvConfig.MigrationsPath,
		config.SrvConfig.DBConnStr,
	)
	if err != nil {
		logger.Log.Fatalln("Error initializing migrate:", err)
	}
	// Apply all available migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Log.Fatalln("Error applying migrations:", err)
	}
	logger.Log.Infoln("Database migrations applied successfully!")
	db.dbconn.Ping()
	return &db
}

// Close DB connection
func (d *DB) Close() error {
	return d.dbconn.Close()
}

// // IsCookieExist return true if cookie is exist in DB
// func (d *DB) IsCookieExist(cookie string) bool {
// 	err := d.dbconn.Ping()
// 	if err != nil {
// 		logger.Log.Warnln(err)
// 	}
// 	var isFound bool
// 	var queryres string
// 	err = d.dbconn.QueryRow("SELECT cookie FROM storage WHERE cookie=$1", cookie).Scan(&queryres)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			isFound = false
// 		} else {
// 			logger.Log.Infoln(err)
// 		}
// 	} else {
// 		if queryres == cookie {
// 			logger.Log.Infoln("cookie is exist in DB", queryres)
// 			isFound = true
// 		}
// 	}
// 	return isFound
// }

// // GetDataByUID return batch of URLs whitch user set.
// func (d *DB) GetDataByUID(uid uint32) []model.ResponseDataBatchByCookie {
// 	var result []model.ResponseDataBatchByCookie

// 	rows, err := d.dbconn.Query("SELECT shortURL, originalURL FROM storage WHERE uuid=$1", uid)
// 	if err != nil {
// 		logger.Log.Fatal(err)
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		var r model.ResponseDataBatchByCookie
// 		if err := rows.Scan(&r.ShortURL, &r.OriginalURL); err != nil {
// 			logger.Log.Fatal(err)
// 		}
// 		result = append(result, r)
// 	}
// 	if err := rows.Err(); err != nil {
// 		logger.Log.Fatal(err)
// 	}
// 	return result
// }

// // IsCreator return true if user is creator of shortURL else false
// func (d *DB) IsCreator(shortURL string, cookie string) bool {
// 	err := d.dbconn.Ping()
// 	if err != nil {
// 		logger.Log.Warnln(err)
// 	}
// 	var resultCookie string
// 	err = d.dbconn.QueryRow("select cookie from storage where shorturl=$1", shortURL).Scan(&resultCookie)
// 	if err != nil {
// 		logger.Log.Warnln("error whele select cookie from DB", err)
// 	}
// 	if resultCookie == cookie {
// 		return true
// 	} else {
// 		return false
// 	}
// }
