package apiserver

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/QMAwerda/http-rest-api/internal/app/model"
	"github.com/QMAwerda/http-rest-api/internal/app/store"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
)

// Тут будет описана более легковесная версия сервера, она не будет знать про запуск сервера, про запуск http сервера, она будет
// только уметь обрабатывать входящий запрос и будет реализовывать интерфейс httpHandler, это поможет прокинуть ее напрямую в функцию
// listen and serv из пакета http, тем самым мы убираем с нее все лишние задачи, потому что, например, в тестах нам не нужно
// запускать http сервер и мы можем эту структуру использовать как есть, без доп методов. А функцию http сервера мы вынесем в
// отдельную функцию start

// выносим в константу, чтоб каждый раз не инициализировать
const (
	sessionName = "randomName"
)

var (
	errIncorrectEmailOrPassword = errors.New("incorrect email or password")
)

type server struct {
	router       *mux.Router
	logger       *logrus.Logger
	store        store.Store
	sessionStore sessions.Store
}

func newServer(store store.Store, sessionStore sessions.Store) *server {
	s := &server{
		router:       mux.NewRouter(),
		logger:       logrus.New(),
		store:        store,
		sessionStore: sessionStore,
	}

	s.configureRouter()

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) configureRouter() {
	s.router.HandleFunc("/users", s.handleUsersCreate()).Methods("POST") // создаем эндпоинт для регистрации пользователей
	s.router.HandleFunc("/sessions", s.handleSessionsCreate()).Methods("POST")
}

// Пример запроса на добавление: http POST http://localhost:8080/users email=user@example.org password=password
// Для регистрации
func (s *server) handleUsersCreate() http.HandlerFunc {
	// создадим структуру с полями, которые нам заполнит пользователь при регистрации
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Функция, которую мы используем и возвращаем для каждого входящего запроса
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		// Декодируем json из Request в нашу структуру
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u := &model.User{
			Email:    req.Email,
			Password: req.Password,
		}
		if err := s.store.User().Create(u); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err) // статус код - пользователь передал некорректные данные
			return
		}

		u.Sanitize()
		s.respond(w, r, http.StatusCreated, u)
	}
}

// Для аутентификации
func (s *server) handleSessionsCreate() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u, err := s.store.User().FindByEmail(req.Email)
		// теперь нужно сравнить переданный пароль с хешем из бд
		// либо в бд нет такого пользователя || либо пароль неверный
		if err != nil || !u.ComparePassword(req.Password) {
			// в качестве ошибки не нужно передавать инфу, что мы не нашли пользователя, потому что
			// злоумышленники смогут брутфорсить и подобрать есть ли такой email в системе или нет
			// поэтому передадим созданную нами абстрактную ошибку, в которой будет говориться:
			// email или пароль указан неверно, а что конкретно - мы не скажем
			s.error(w, r, http.StatusUnauthorized, errIncorrectEmailOrPassword)
			return
		}
		// после аутентификации пользователя мы должны выдать ему jwt токен или сессионные куки, с которыми он сможет
		// ходить на наши защищенные ресурсы
		// в данном случае после аутентификации мы передадим заголовок сет куки с его сессией
		session, err := s.sessionStore.Get(r, sessionName)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		// Добавим ID, чтобы потом в middleware проверять его и если нет возвращать неавторизован и тп
		session.Values["user_id"] = u.ID
		if err := s.sessionStore.Save(r, w, session); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, nil)

	}
}

// хелпер для более удобного вывода ошибки, полученной в процессе работы нашего обработчика
func (s *server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
	// в мапе сохраним значение нашей ошибки err.Error()
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code) // запишем в наш ответ статус код
	if data != nil {    // если передаем и данные
		json.NewEncoder(w).Encode(data) // то сериализуем их
	}
}
