package handler

import (
	"net/http"

	rulesvc "streampass/backend/internal/application/rule"
	ruledomain "streampass/backend/internal/domain/rule"
	httpx "streampass/backend/internal/infrastructure/http"
)

// RuleHandler exposes GET /rules and POST /rules (admin).
type RuleHandler struct {
	svc *rulesvc.Service
}

// NewRuleHandler builds the Rule Service HTTP handler.
func NewRuleHandler(svc *rulesvc.Service) *RuleHandler {
	return &RuleHandler{svc: svc}
}

type ruleDTO struct {
	Kind    string `json:"kind"`
	Pattern string `json:"pattern"`
	Mode    string `json:"mode"`
}

type ruleSetResponse struct {
	Version   int       `json:"version"`
	Rules     []ruleDTO `json:"rules"`
	CreatedAt string    `json:"created_at"`
}

// GetLatest handles "GET /rules".
func (h *RuleHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	set, err := h.svc.GetLatest(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toRuleSetResponse(set))
}

type publishRulesRequest struct {
	Rules []ruleDTO `json:"rules"`
}

// Publish handles "POST /rules" (admin-only, gated by RequireAdminKey in
// the router).
func (h *RuleHandler) Publish(w http.ResponseWriter, r *http.Request) {
	var req publishRulesRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	rules := make([]ruledomain.Rule, len(req.Rules))
	for i, dto := range req.Rules {
		rules[i] = ruledomain.Rule{Kind: ruledomain.Kind(dto.Kind), Pattern: dto.Pattern, Mode: ruledomain.Mode(dto.Mode)}
	}

	set, err := h.svc.Publish(r.Context(), rules)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toRuleSetResponse(set))
}

func toRuleSetResponse(set *ruledomain.Set) ruleSetResponse {
	dtos := make([]ruleDTO, len(set.Rules))
	for i, rl := range set.Rules {
		dtos[i] = ruleDTO{Kind: string(rl.Kind), Pattern: rl.Pattern, Mode: string(rl.Mode)}
	}
	return ruleSetResponse{Version: set.Version, Rules: dtos, CreatedAt: set.CreatedAt.Format(httpx.TimeFormat)}
}
