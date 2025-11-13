package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Parnishkaspb/avito/internal/config"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	port   int
	host   string
	router *http.ServeMux
}

func (s Server) createTeamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		tmp := fmt.Sprintf("Ошибка обращения к createTeamHandler. Метод: %s - требуемый: POST", r.Method)
		log.Fatal(tmp)
		return
	}

	var teamAdd requestTeamAdd

	if err := json.NewDecoder(r.Body).Decode(&teamAdd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(teamAdd)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := map[string]interface{}{
		"team": teamAdd,
	}
	json.NewEncoder(w).Encode(response)
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

func New(serverConfig config.ServerConfig) IServer {
	return &Server{
		port:   serverConfig.Port,
		host:   serverConfig.Host,
		router: http.NewServeMux(),
	}
}
