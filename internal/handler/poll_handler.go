package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"polling-service/internal/domain"
	"polling-service/internal/service"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type PollHandler struct {
	pollService *service.PollService
	voteService *service.VoteService
}

func NewPollHandler(pollService *service.PollService, voteService *service.VoteService) *PollHandler {
	return &PollHandler{pollService: pollService, voteService: voteService}
}

func (h *PollHandler) CreatePoll(w http.ResponseWriter, r *http.Request) {
	var req domain.CreatePollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	poll, err := h.pollService.CreatePoll(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(poll)
}

func (h *PollHandler) GetPoll(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		slog.Error("Empty poll ID")
		http.Error(w, "Invalid poll ID", http.StatusBadRequest)
		return
	}

	poll, err := h.pollService.GetPoll(r.Context(), id)
	if err != nil {
		if err == service.ErrPollNotFound {
			http.Error(w, "Poll not found", http.StatusNotFound)
			return
		}
		slog.Error("%w", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(poll)
}

func (h *PollHandler) ListPolls(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	polls, err := h.pollService.ListPolls(r.Context(), page, pageSize)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(polls)
}

func (h *PollHandler) ListVotes(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	votes, err := h.voteService.ListVotes(r.Context(), page, pageSize)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		slog.Error("Internal server error: %w", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(votes)
}

func (h *PollHandler) Vote(w http.ResponseWriter, r *http.Request) {
	pollId := chi.URLParam(r, "id")
	if pollId == "" {
		slog.Error("Empty poll ID")
		http.Error(w, "Invalid poll ID", http.StatusBadRequest)
		return
	}
	var req domain.VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.OptionID == "" {
		http.Error(w, "option_id is required", http.StatusBadRequest)
	}
	if req.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
	}
	poll, err := h.voteService.Vote(r.Context(), pollId, req.OptionID, req.UserID)
	if err != nil {
		switch err {
		case service.ErrPollNotFound:
			http.Error(w, "Poll not found", http.StatusNotFound)
		case domain.ErrInvalidOption:
			http.Error(w, "Invalid option", http.StatusBadRequest)
		case domain.ErrAlreadyVoted:
			http.Error(w, "User already voted in this poll", http.StatusConflict)
		default:
			slog.Error("Vote failed", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(poll)

}
