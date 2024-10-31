package util

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/atomic"

	"github.com/evolutionlandorg/evo-backend/util/log"
	"github.com/gin-gonic/gin"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
)

var (
	db     *gorm.DB
	dbHost = GetEnv("MYSQL_HOST", "127.0.0.1")
	dbUser = GetEnv("MYSQL_USER", "root")
	dbPass = GetEnv("MYSQL_PASS", "123456")
	dbName = GetEnv("MYSQL_DB", "consensus-backend")
	dbPort = GetEnv("MYSQL_PORT", "3306")
)

type GormDB struct {
	*gorm.DB
	ctx     context.Context
	gdbDone *atomic.Bool
}

func init() {
	if gin.Mode() == "test" {
		dbName = "consensus-backend-test"
	}
}

type Logger interface {
	Print(v ...interface{})
}

func WithContextDb(ctx context.Context) *gorm.DB {
	return db
}

func CreateDatabaseIfNotExist(host, port, dbname, user, password string) error {
	// 构建连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port)

	// 连接到数据库
	sqltrace.Register("mysql", &mysqlDriver.MySQLDriver{})
	tdb, err := gorm.Open("mysql", dsn)
	if err != nil {
		return err
	}

	// 检查数据库是否存在
	var result int64
	err = tdb.Where("schema_name = ?", dbname).Table("information_schema.schemata").Count(&result).Error
	if err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	if result == 0 {
		// 数据库不存在，创建数据库
		return tdb.Exec(fmt.Sprintf("CREATE DATABASE `%s`", dbname)).Error
	}
	return nil
}

func InitMysql(log ...Logger) error {
	// check the dbName is exist, if not exist create it
	if err := CreateDatabaseIfNotExist(dbHost, dbPort, dbName, dbUser, dbPass); err != nil {
		return err
	}

	sqltrace.Register("mysql", &mysqlDriver.MySQLDriver{})
	tdb, err := gorm.Open("mysql",
		dbUser+":"+dbPass+"@tcp("+dbHost+":"+dbPort+")/"+dbName+"?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		return err
	}
	tdb.DB().SetMaxIdleConns(10)
	tdb.DB().SetMaxOpenConns(100)
	tdb.DB().SetConnMaxLifetime(5 * time.Minute)
	if len(log) != 0 && log[0] != nil {
		tdb.SetLogger(log[0])
		tdb.LogMode(true)
	}

	db = tdb
	return db.DB().Ping()
}

func DbBegin(ctx context.Context) *GormDB {
	db := WithContextDb(ctx)
	txn := db.Begin()
	// Panic(db.Begin().Error, "DbBegin error")
	return &GormDB{txn, ctx, atomic.NewBool(false)}
}

func (c *GormDB) Context() context.Context {
	return c.ctx
}

func (c *GormDB) DbCommit() {
	if c.gdbDone.Load() {
		return
	}
	tx := c.Commit()
	c.gdbDone.Store(true)
	if err := tx.Error; err != nil && err != sql.ErrTxDone {
		log.Debug("Fatal error DbCommit: %s", err)
	}
}

func (c *GormDB) DbRollback() {
	if c.gdbDone.Load() {
		return
	}
	tx := c.Rollback()
	c.gdbDone.Store(true)
	if err := tx.Error; err != nil && err != sql.ErrTxDone {
		log.Debug("Fatal error DbRollback: %s", err)
	}
}

func TruncateTable() {
	c, _ := OpenTestConnection()
	if gin.Mode() == "test" {
		c.Exec("TRUNCATE TABLE members")
		c.Exec("TRUNCATE TABLE accounts")
		c.Exec("TRUNCATE TABLE invites")
		c.Exec("TRUNCATE TABLE withdraws")
	}
}

func OpenTestConnection() (db *gorm.DB, err error) {
	dbDSN := "root" + ":" + "" + "@tcp(" + "127.0.0.1" + ")/" + "consensus-backend-test" + "?charset=utf8&parseTime=True&loc=Local"
	db, err = gorm.Open("mysql", dbDSN)
	return
}
