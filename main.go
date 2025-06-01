package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"embarcados/internal/database"
	"embarcados/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Configuração do pool de conexões
	pgxConfig, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Erro ao analisar a configuração do banco de dados: %v\n", err)
	}
	pgxConfig.MaxConns = 100
	pgxConfig.MinConns = 5
	pgxConfig.MaxConnLifetime = time.Hour
	pgxConfig.HealthCheckPeriod = 5 * time.Minute
	pgxConfig.MaxConnIdleTime = 15 * time.Minute
	pgxConfig.ConnConfig.ConnectTimeout = 5 * time.Second

	pool, err := pgxpool.NewWithConfig(context.Background(), pgxConfig)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v\n", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Erro ao se comunicar com o banco de dados: %v\n", err)
	}

	db := database.New(pool)
	h := handler.NewHandler(db)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Rotas para criar registros (POST)
	r.Post("/volume", h.VolumeHandler)
	r.Post("/vazao", h.FlowRateHandler)

	// Rotas para leitura (GET)
	r.Get("/volume-periodo", h.VolumeByPeriodHandler)
	r.Get("/volumes", h.GetVolumes)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("PORT não definida, usando porta padrão %s\n", port)
	}

	addr := ":" + port
	log.Printf("Servidor escutando na porta %s\n", port)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
