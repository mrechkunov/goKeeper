package repository

import (
	"database/sql"
	"errors"

	"github.com/mrechkunov/goKeeper.git/internal/logger"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type DB struct {
	dbconn  *sql.DB
	counter int
}

// NewDB set new connection to DB from config, applying all migrations, set counter of rows
func NewDB() *DB {
	db, err := NewConnect()
	if err != nil {
		logger.Log.Errorln("error while db connection")
	}

	m, err := migrate.New(
		config.ConfigAdreses.MigrationsPath,
		config.ConfigAdreses.DBConnStr,
	)
	if err != nil {
		logger.Log.Fatalln("Error initializing migrate:", err)
	}
	// Apply all available migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Log.Fatalln("Error applying migrations:", err)
	}
	logger.Log.Infoln("Database migrations applied successfully!")
	db.Ping()

	// set DB struct (counter uuid & db conn)
	var cnt int
	_ = db.QueryRow("SELECT count FROM storage ORDER BY count DESC LIMIT 1").Scan(
		&cnt)
	cnt++
	var retDB = &DB{
		dbconn:  db,
		counter: cnt,
	}
	return retDB
}

// SetData insert to DB new row with shortURL, originalURL, UID
func (d *DB) SetData(shortURL string, originalURL string, cookie string) error {
	err := d.dbconn.Ping()
	if err != nil {
		logger.Log.Fatal(err)
	}
	// проверяем есть ли такой URL в DB
	var shortURLFromDB string
	d.dbconn.QueryRow("SELECT shorturl FROM storage WHERE shorturl=$1", shortURL).Scan(&shortURLFromDB)
	if shortURLFromDB == shortURL {
		logger.Log.Infoln("shortURL already exist in DB")
		return errors.New("409 Conflict")
	} else {
		uid, _ := cryptoauth.GetIDFromCookie(cookie)
		sqlStatement := `INSERT INTO storage 
			(count, uuid, originalurl, shorturl, cookie, isdeleted) 
			VALUES ($1, $2, $3, $4, $5, $6)`
		_, err := d.dbconn.Exec(sqlStatement, d.counter, uid, originalURL, shortURL, cookie, false)
		if err != nil {
			logger.Log.Errorln("error while insert to db", err)
			return err
		}
		d.counter++
	}
	return nil
}

// Return originalURL from DB by shortURL if row is not exist, return isFound = false
func (d *DB) GetData(shortURL string) (originalURL string, isFound bool) {
	err := d.dbconn.Ping()
	if err != nil {
		logger.Log.Warnln(err)
	}
	err = d.dbconn.QueryRow("SELECT originalurl FROM storage WHERE shorturl=$1", shortURL).Scan(&originalURL)
	isFound = true
	if err != nil {
		isFound = false
	}
	return originalURL, isFound
}

// Close DB connection
func (d *DB) Close() error {
	return d.dbconn.Close()
}

// IsCookieExist return true if cookie is exist in DB
func (d *DB) IsCookieExist(cookie string) bool {
	err := d.dbconn.Ping()
	if err != nil {
		logger.Log.Warnln(err)
	}
	var isFound bool
	var queryres string
	err = d.dbconn.QueryRow("SELECT cookie FROM storage WHERE cookie=$1", cookie).Scan(&queryres)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			isFound = false
		} else {
			logger.Log.Infoln(err)
		}
	} else {
		if queryres == cookie {
			logger.Log.Infoln("cookie is exist in DB", queryres)
			isFound = true
		}
	}
	return isFound
}

// GetDataByUID return batch of URLs whitch user set.
func (d *DB) GetDataByUID(uid uint32) []model.ResponseDataBatchByCookie {
	var result []model.ResponseDataBatchByCookie

	rows, err := d.dbconn.Query("SELECT shortURL, originalURL FROM storage WHERE uuid=$1", uid)
	if err != nil {
		logger.Log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var r model.ResponseDataBatchByCookie
		if err := rows.Scan(&r.ShortURL, &r.OriginalURL); err != nil {
			logger.Log.Fatal(err)
		}
		result = append(result, r)
	}
	if err := rows.Err(); err != nil {
		logger.Log.Fatal(err)
	}
	return result
}

// IsDeleted return true if shortURL is mark as deleted
func (d *DB) IsDeleted(shortURL string) bool {
	err := d.dbconn.Ping()
	if err != nil {
		logger.Log.Warnln(err)
	}
	var result bool
	err = d.dbconn.QueryRow("SELECT isdeleted FROM storage WHERE shorturl=$1", shortURL).Scan(&result)
	if err != nil {
		logger.Log.Warnln(err)
	}
	return result
}

// IsCreator return true if user is creator of shortURL else false
func (d *DB) IsCreator(shortURL string, cookie string) bool {
	err := d.dbconn.Ping()
	if err != nil {
		logger.Log.Warnln(err)
	}
	var resultCookie string
	err = d.dbconn.QueryRow("select cookie from storage where shorturl=$1", shortURL).Scan(&resultCookie)
	if err != nil {
		logger.Log.Warnln("error whele select cookie from DB", err)
	}
	if resultCookie == cookie {
		return true
	} else {
		return false
	}
}

// SetIsDeleted  mark all shortURLs from slice as deleted
func (d *DB) SetIsDeleted(shortURLs []string) {
	err := d.dbconn.Ping()
	if err != nil {
		logger.Log.Warnln(err)
	}
	sqlStatement := `UPDATE storage AS s
		SET isdeleted = true
		FROM (SELECT * FROM UNNEST($1::text[]) AS t(shorturl)) AS data
		WHERE s.shorturl = data.shorturl;`
	_, err = d.dbconn.Exec(sqlStatement, shortURLs)
	if err != nil {
		logger.Log.Errorln("error while UPDATE isdeleted fiels in db", err)
	}
}

func (d *DB) GetStatData() (statData model.ResponseStatData) {
	err := d.dbconn.QueryRow("SELECT COUNT(DISTINCT shorturl) FROM storage;").Scan(&statData.Urls)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			statData.Urls = 0
		} else {
			logger.Log.Infoln(err)
		}
	}
	err = d.dbconn.QueryRow("SELECT COUNT(DISTINCT cookie) FROM storage;").Scan(&statData.Users)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			statData.Users = 0
		} else {
			logger.Log.Infoln(err)
		}
	}
	return
}
