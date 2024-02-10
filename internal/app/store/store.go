package store

import (
	"database/sql"

	_ "github.com/lib/pq" // анонимный импорт, чтобы методы в код не импортировались
)

type Store struct {
	config         *Config
	db             *sql.DB
	UserRepository *UserRepository
}

func New(config *Config) *Store {
	return &Store{
		config: config,
	}
}

// Метод Open может вернуть ошибку, используется при инициализации хранилища,
// при попытке подключиться к БД, возможно поймает ошибку и зафейлится на этом
func (s *Store) Open() error {
	db, err := sql.Open("postgres", s.config.DatabaseURL) // (драйвер, адрес бд)
	if err != nil {
		return err
	}
	// sql.Open() не создает реальное соединение. Оно создается, когда происходит первый запрос
	// db.Ping() делает SELECT запрос строки к БД, так мы проверяем, что все работает
	if err := db.Ping(); err != nil {
		return err
	}

	s.db = db

	return nil
}

// Этот метод нужен, чтобы отключиться от базы данных, когда сервер закончил свою работу
func (s *Store) Close() {
	s.db.Close()
}

// store.User().Create() - пример обращения к репозиторию из внешнего мира
func (s *Store) User() *UserRepository {
	if s.UserRepository != nil {
		return s.UserRepository
	}

	s.UserRepository = &UserRepository{
		store: s,
	}

	return s.UserRepository
}
