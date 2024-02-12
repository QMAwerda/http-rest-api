package apiserver

import (
	"io"
	"net/http"

	"github.com/QMAwerda/http-rest-api/internal/app/store"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// APIServer
type APIServer struct {
	config *Config
	logger *logrus.Logger
	router *mux.Router
	store  *store.Store
}

// New
// В этом методе мы инициализируем наш api сервер
func New(config *Config) *APIServer {
	return &APIServer{
		config: config,
		logger: logrus.New(),
		router: mux.NewRouter(),
	}
}

// Start
// С помощью этой функции мы будем запускать http сервер, подключаться к БД и тд.
// Эта функция может вернуть ошибку (например, если занят порт и тд)
func (s *APIServer) Start() error {
	if err := s.configureLogger(); err != nil {
		return err
	}

	s.configureRouter()

	if err := s.configureStore(); err != nil {
		return err
	}

	s.logger.Info("starting api server")

	return http.ListenAndServe(s.config.BindAddr, s.router)
}

func (s *APIServer) configureLogger() error {
	level, err := logrus.ParseLevel(s.config.LogLevel)
	if err != nil {
		return err
	}

	s.logger.SetLevel(level)

	return nil
}

func (s *APIServer) configureRouter() {
	s.router.HandleFunc("/hello", s.handleHello())
}

func (s *APIServer) configureStore() error {
	st := store.New(s.config.Store)
	if err := st.Open(); err != nil {
		return err
	}

	s.store = st
	return nil
}

func (s *APIServer) handleHello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello")
	}
}
