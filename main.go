package main

import (
	"context"
	"embarcados/internal/database"
	"embarcados/internal/handler"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	pgxConfig, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Erro ao analisar a configuração do banco de dados: %v\n", err)
		panic(err)
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
		panic(err)
	}
	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatalf("Erro ao se comunicar com o banco de dados: %v\n", err)
		panic(err)
	}
	defer pool.Close()

	db := database.New(pool)
	handler := handler.NewHandler(db)

	http.HandleFunc("POST /volume", handler.VolumeHandler)
	http.HandleFunc("POST /vazao", handler.FlowRateHandler)
	http.HandleFunc("GET /volume-periodo", handler.VolumeByPeriodHandler)
	http.HandleFunc("GET /volume", handler.GetVolumes)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("PORT não definida, usando porta padrão %s\n", port)
	}

	addr := ":" + port
	log.Printf("Servidor escutando na porta %s\n", port)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
