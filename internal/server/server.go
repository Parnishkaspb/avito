package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Parnishkaspb/avito/internal/jwt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Parnishkaspb/avito/internal/config"
	"github.com/Parnishkaspb/avito/internal/constants"
	"github.com/Parnishkaspb/avito/internal/database"
	"github.com/Parnishkaspb/avito/internal/models"
)

type Server struct {
	port       int
	host       string
	router     *http.ServeMux
	db         database.DB
	jwtService *jwt.Service
}

func New(serverConfig config.ServerConfig, db database.DB, jwt_secret string) *Server {
	jwtConfig := jwt.Config{
		SecretKey:       jwt_secret,
		Issuer:          "avito",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}

	return &Server{
		port:       serverConfig.Port,
		host:       serverConfig.Host,
		router:     http.NewServeMux(),
		db:         db,
		jwtService: jwt.New(jwtConfig),
	}
}

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		claims, err := s.jwtService.ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "userID", claims.UserID)
		ctx = context.WithValue(ctx, "userEmail", claims.Name)
		r = r.WithContext(ctx)

		next(w, r)
	}
}

func (s *Server) createTeamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		tmp := fmt.Sprintf("Ошибка обращения к createTeamHandler. Метод: %s - требуемый: POST", r.Method)
		log.Fatal(tmp)
		return
	}

	var teamAdd models.RequestTeamAdd

	if err := json.NewDecoder(r.Body).Decode(&teamAdd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	exists, err := s.db.CreateTeam(context.Background(), teamAdd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := map[string]interface{}{
			"team": teamAdd,
		}
		json.NewEncoder(w).Encode(response)
	} else {
		s.writeError(w, constants.TEAM_EXISTS, "team_name already exists", http.StatusBadRequest)
	}

}

func (s *Server) getTeamHandler(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) setupRoutes() {
	s.router.HandleFunc("POST /team/add", s.createTeamHandler)
}

func (s *Server) RunServer(ctx context.Context) error {
	s.setupRoutes()

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(s.port),
		Handler: s.router,
	}

	started := make(chan bool, 1)
	serverErr := make(chan error, 1)

	go func() {
		log.Printf("Запуск сервера на http://%s:%d", s.host, s.port)
		started <- true

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- fmt.Errorf("ошибка сервера: %w", err)
		}
	}()

	<-started
	log.Println("Сервер успешно запущен")

	select {
	case <-ctx.Done():
		log.Println("Получен сигнал остановки...")
		return s.gracefulShutdown(server)
	case err := <-serverErr:
		return err
	}
}

func (s *Server) gracefulShutdown(server *http.Server) error {

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Ошибка: %v", err)
		return server.Close()
	}

	log.Println("Сервер остановлен")
	return nil
}

func (s *Server) writeError(w http.ResponseWriter, code, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": msg,
		},
	})
}
