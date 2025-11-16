package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Parnishkaspb/avito/internal/helper"
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

type contextKey string

const (
	userIDKey   contextKey = "userID"
	userNameKey contextKey = "userName"
	userRoleKey contextKey = "userRole"
)

func getUserRole(ctx context.Context) (bool, bool) {
	role, ok := ctx.Value(userRoleKey).(bool)
	return role, ok
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
		ctx = context.WithValue(ctx, userIDKey, claims.UserID)
		ctx = context.WithValue(ctx, userNameKey, claims.Name)
		ctx = context.WithValue(ctx, userRoleKey, claims.Role)
		r = r.WithContext(ctx)

		next(w, r)
	}
}

func (s *Server) adminRoleMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		role, _ := getUserRole(r.Context())

		if !role {
			s.writeError(w, constants.NOT_FOUND, "resource not found", http.StatusUnauthorized)
			return
		}

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

	var teamAdd models.RequestTeamAddResponse

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
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		http.Error(w, "team_name обязателен!", http.StatusBadRequest)
		return
	}

	teamID, found, err := s.db.ReturnTeamID(context.Background(), teamName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !found {
		s.writeError(w, constants.NOT_FOUND, "resource not found", http.StatusNotFound)
		return
	}

	teamMembers, err := s.db.ReturnTeamMembersByTeamID(context.Background(), teamID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	teamRequest := models.RequestTeamAddResponse{
		TeamName: teamName,
		Members:  teamMembers,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(teamRequest)
}

func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.ID == "" || req.Name == "" {
		http.Error(w, "Отсутствует из параметров обязательных параметров", http.StatusBadRequest)
		return
	}

	role, err := s.db.CheckRoleUser(context.Background(), req.ID)

	tokenPair, err := s.jwtService.GenerateTokenPair(req.ID, req.Name, role)
	if err != nil {
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenPair)
}

func (s *Server) setIsActiveUserHandler(w http.ResponseWriter, r *http.Request) {
	var userActive models.UserActive

	if err := json.NewDecoder(r.Body).Decode(&userActive); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	exists, err := s.db.CheckUser(context.Background(), userActive.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if !exists {
		s.writeError(w, constants.NOT_FOUND, "resource not found", http.StatusNotFound)
		return
	}

	_, err = s.db.UpdateActive(context.Background(), userActive.UserID, userActive.IsActive)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	info, err := s.db.GetUser(context.Background(), userActive.UserID)

	response := map[string]models.UserActiveResponse{
		"user": info,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) getReviewHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id обязателен!", http.StatusBadRequest)
		return
	}

	pullRequests, err := s.db.ReturnUserReviewByUserID(context.Background(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	response := models.UserPullRequestsResponse{
		UserID:       userID,
		PullRequests: pullRequests,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) createPullRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		tmp := fmt.Sprintf("Ошибка обращения к createTeamHandler. Метод: %s - требуемый: POST", r.Method)
		log.Fatal(tmp)
		return
	}

	var PRCR models.PullRequestCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&PRCR); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	exists, err := s.db.CheckPR(context.Background(), PRCR.PullRequestId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if exists {
		s.writeError(w, constants.PR_EXISTS, "PR id already exists", http.StatusConflict)
		return
	}

	info, err := s.db.GetUser(context.Background(), PRCR.AuthorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if info.UserID == "" || info.TeamName == "" {
		s.writeError(w, constants.NOT_FOUND, "resource not found", http.StatusNotFound)
		return
	}

	//TODO:
	// 1) Посмотреть участников в команде исключая автора и все должны быть активными
	// 2) Создать PR
	// 3) Создать Коннект между участниками и Пулл

	teamMates, err := s.db.ReturnTeamMembersByUserID(context.Background(), info.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pullRequest, err := s.db.CreatePullRequest(context.Background(), PRCR.PullRequestId, PRCR.PullRequestName, PRCR.AuthorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	randomTeamMates := helper.PickRandomTeamMates(teamMates, 2)

	_, err = s.db.CreatePullRequestAssignedReview(context.Background(), pullRequest.PullRequestID, randomTeamMates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	response := models.PullRequestResponse{
		PullRequestID:     pullRequest.PullRequestID,
		PullRequestName:   pullRequest.PullRequestName,
		AuthorID:          pullRequest.AuthorID,
		Status:            pullRequest.Status,
		AssignedReviewers: randomTeamMates,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) setupRoutes() {
	s.router.HandleFunc("POST /team/add", s.createTeamHandler)
	s.router.HandleFunc("GET  /team/get", s.authMiddleware(s.getTeamHandler))
	s.router.HandleFunc("POST /login", s.loginHandler)
	s.router.HandleFunc("POST /users/setIsActive", s.authMiddleware(s.adminRoleMiddleware(s.setIsActiveUserHandler)))
	s.router.HandleFunc("GET  /users/getReview", s.authMiddleware(s.getReviewHandler))
	s.router.HandleFunc("POST /pullRequest/create", s.authMiddleware(s.adminRoleMiddleware(s.createPullRequestHandler)))

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
