package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/QMAwerda/http-rest-api/internal/app/model"
	"github.com/QMAwerda/http-rest-api/internal/app/store"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
)

//Пример запроса, после активации ./apiserver (в другом терминале)
// http -v --session=user http://localhost:8080/private/whoami "Origin: google.com"

// Тут будет описана более легковесная версия сервера, она не будет знать про запуск сервера, про запуск http сервера, она будет
// только уметь обрабатывать входящий запрос и будет реализовывать интерфейс httpHandler, это поможет прокинуть ее напрямую в функцию
// listen and serv из пакета http, тем самым мы убираем с нее все лишние задачи, потому что, например, в тестах нам не нужно
// запускать http сервер и мы можем эту структуру использовать как есть, без доп методов. А функцию http сервера мы вынесем в
// отдельную функцию start

// выносим в константу, чтоб каждый раз не инициализировать
const (
	sessionName        = "randomName"
	ctxKeyUser  ctxKey = iota
	ctxKeyRequestID
)

var (
	errIncorrectEmailOrPassword = errors.New("incorrect email or password")
	errNotAuthenticated         = errors.New("not authenticated")
)

type ctxKey int8 // создаем тип ключа для контекста

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
	s.router.Use(s.setRequestID)
	s.router.Use(s.logRequest)
	//AllowedOrigins разрешает доступ с определенных источников, а "*" показывает, что разрешает с любых доменов
	//Теперь мы сможем отправлять запросы к нашему серверу из браузера
	s.router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))

	s.router.HandleFunc("/users", s.handleUsersCreate()).Methods("POST") // создаем эндпоинт для регистрации пользователей
	s.router.HandleFunc("/sessions", s.handleSessionsCreate()).Methods("POST")

	// выделим подроутер, запросы на эти middleware будут доступны так:
	// /private/... (роутеры выше не прикрыты middleware и они доступны для гостей и любых запросов)
	private := s.router.PathPrefix("/private").Subrouter()
	private.Use(s.authenticateUser)
	private.HandleFunc("/whoami", s.handleWhoami()).Methods("GET")
}

func (s *server) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// локальный логгер с параметрами характерными только для конктерного запроса
		logger := s.logger.WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
			"request_id":  r.Context().Value(ctxKeyRequestID),
		})
		// started Get /endpoint...
		logger.Infof("started %s %s", r.Method, r.RequestURI)

		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		logger.Infof(
			"completed with %d %s in %v", // статус код, его текстовое представление in время
			rw.code,
			http.StatusText(rw.code),
			time.Since(start),
		)
	})
}

func (s *server) authenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// получим название сессии
		session, err := s.sessionStore.Get(r, sessionName)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		id, ok := session.Values["user_id"]
		if !ok { // если в нашей сессии нет айди пользователя
			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
			return
		}
		// если айди все таки есть, но пользователь не найден
		u, err := s.store.User().Find(id.(int))
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
			return
		}
		// чтобы не создавать каждый раз пользователя, мы можем использовать контекст, в котором передадим пользователя
		// также контекст используется для распространения информации о том, что какое-то действие можно завершать
		// context.WithValue(родительский контекст, ключ куда запишем значение)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyUser, u)))
	})
}

func (s *server) handleWhoami() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, r, http.StatusOK, r.Context().Value(ctxKeyUser).(*model.User))
	}
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
