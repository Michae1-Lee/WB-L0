package postgres

import (
	"database/sql"
	"embed"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"os"
)

//go:embed migrations/*.sql
var migrations embed.FS

func InitDB(logger *zap.SugaredLogger) (*sqlx.DB, error) {
	dsn := os.Getenv("PG_DSN")

	logger.Infof("Connecting to DB with URL: %s", dsn)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, errors.WithMessage(err, "init db failed")
	}

	if err := runMigrations(logger, db.DB, "migrations"); err != nil {
		return nil, errors.WithMessage(err, "migrations failed")
	}

	return db, nil
}

func runMigrations(logger *zap.SugaredLogger, db *sql.DB, migrationsDir string) error {
	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return errors.Wrap(err, "set dialect failed")
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		return errors.Wrap(err, "apply migrations failed")
	}

	logger.Info("Migrations applied successfully")
	return nil
}
