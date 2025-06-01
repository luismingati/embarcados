package handler

import (
	"context"
	"embarcados/internal/database"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Handler struct {
	db *database.Queries
}

func NewHandler(db *database.Queries) *Handler {
	return &Handler{
		db: db,
	}
}

type ReqBody struct {
	Value string `json:"value"`
}

func (h *Handler) VolumeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req ReqBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Erro ao ler o corpo da requisição: %v\n", err)
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	valueFloat, err := strconv.ParseFloat(req.Value, 64)
	if err != nil {
		log.Printf("Erro ao parsear string para float: '%s', erro: %v\n", req.Value, err)
		http.Error(w, "Payload inválido: 'value' deve ser string de número (ex: \"123.45\")", http.StatusBadRequest)
		return
	}

	params := database.CreateVolumeParams{
		Value: valueFloat,
		CreatedAt: pgtype.Timestamp{
			Time:  time.Now(),
			Valid: true,
		},
	}
	created, err := h.db.CreateVolume(context.Background(), params)
	if err != nil {
		log.Printf("Erro ao inserir volume no banco: %v\n", err)
		http.Error(w, "Erro ao salvar volume", http.StatusInternalServerError)
		return
	}

	log.Printf("Volume inserido: ID=%d, Value=%.2f, CreatedAt=%s\n",
		created.ID, created.Value, created.CreatedAt.Time.Format(time.RFC3339),
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *Handler) FlowRateHandler(w http.ResponseWriter, r *http.Request) {
	var req ReqBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Erro ao ler o corpo da requisição: %v\n", err)
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	valueFloat, err := strconv.ParseFloat(req.Value, 64)
	if err != nil {
		log.Printf("Erro ao parsear string para float: '%s', erro: %v\n", req.Value, err)
		http.Error(w, "Payload inválido: 'value' deve ser string de número (ex: \"123.45\")", http.StatusBadRequest)
		return
	}

	params := database.CreateFlowRateParams{
		Value: valueFloat,
		CreatedAt: pgtype.Timestamp{
			Time:  time.Now(),
			Valid: true,
		},
	}
	created, err := h.db.CreateFlowRate(context.Background(), params)
	if err != nil {
		log.Printf("Erro ao inserir vazão no banco: %v\n", err)
		http.Error(w, "Erro ao salvar vazão", http.StatusInternalServerError)
		return
	}

	log.Printf("Vazão inserida: ID=%d, Value=%.2f, CreatedAt=%s\n",
		created.ID, created.Value, created.CreatedAt.Time.Format(time.RFC3339),
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

type VolumePediodReq struct {
	period string
	limit  int32
}

func getLimitByPeriod(period string) (int32, error) {
	switch period {
	case "second":
		return 3600, nil
	case "minute":
		return 60, nil
	case "hour":
		return 24, nil
	case "day":
		return 30, nil
	case "week":
		return 21, nil
	case "month":
		return 12, nil
	case "year":
		return 10, nil
	default:
		return 0, fmt.Errorf("período inválido")
	}
}

type moneyRes struct {
	Period     time.Time `json:"period"`
	TotalValue string    `json:"total_value"`
}

func formatMoneyOutput(res []database.GetVolumesByPeriodRow) []moneyRes {
	out := make([]moneyRes, len(res))

	for i, row := range res {
		valorFormatado := fmt.Sprintf("R$ %.2f", row.TotalValue)
		out[i] = moneyRes{
			Period:     row.Period.Time,
			TotalValue: valorFormatado,
		}
	}

	return out
}

func (h *Handler) GetVolumes(w http.ResponseWriter, r *http.Request) {
	vol, err := h.db.GetVolumes(context.Background(), 100)
	if err != nil {
		log.Printf("Erro ao buscar volumes no banco: %v\n", err)
		http.Error(w, "Erro ao buscar volumes", http.StatusInternalServerError)
		return
	}
	if len(vol) == 0 {
		log.Printf("Nenhum volume encontrado\n")
		http.Error(w, "Nenhum volume encontrado", http.StatusNotFound)
		return
	}
	log.Printf("Volumes encontrados: %d\n", len(vol))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vol)
}

func (h *Handler) VolumeByPeriodHandler(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	typ := r.URL.Query().Get("type")
	fmt.Println(typ)
	if typ == "" {
		typ = "liters"
	} else {
		if typ != "liters" && typ != "money" {
			log.Printf("Tipo inválido: %s\n", typ)
			http.Error(w, "Tipo inválido", http.StatusBadRequest)
			return

		}
	}
	limit, err := getLimitByPeriod(period)
	if err != nil {
		log.Printf("Erro ao obter limite: %v\n", err)
		http.Error(w, "Período inválido", http.StatusBadRequest)
		return
	}

	res, err := h.db.GetVolumesByPeriod(context.Background(), database.GetVolumesByPeriodParams{
		DateTrunc: period,
		Limit:     int32(limit),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("Nenhum volume encontrado para o período %s\n", period)
			http.Error(w, "Nenhum volume encontrado", http.StatusNotFound)
			return
		}
		log.Printf("Erro ao buscar volumes no banco: %v\n", err)
		http.Error(w, "Erro ao buscar volumes", http.StatusInternalServerError)
		return
	}
	if len(res) == 0 {
		log.Printf("Nenhum volume encontrado para o período %s\n", period)
		http.Error(w, "Nenhum volume encontrado", http.StatusNotFound)
		return
	}
	if typ == "money" {
		fmt.Println(typ)
		for i := range res {
			res[i].TotalValue *= 2
		}
		resp := formatMoneyOutput(res)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
		return
	}
	slog.Info("Volumes encontrados", "period", period, "count", len(res))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
