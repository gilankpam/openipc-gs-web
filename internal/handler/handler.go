package handler

import (
	"encoding/json"
	"net/http"

	"github.com/openipc/ezconfig/internal/models"
	"github.com/openipc/ezconfig/internal/service"
)

type Handler struct {
	service *service.ConfigService
}

func NewHandler(svc *service.ConfigService) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) GetRadio(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.GetRadioSettings()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(settings)
}

func (h *Handler) UpdateRadio(w http.ResponseWriter, r *http.Request) {
	var settings models.RadioSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateRadioSettings(&settings); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ... similar handlers for Video, Camera, Telemetry, Alink ...

func (h *Handler) GetVideo(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.GetVideoSettings()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(settings)
}

func (h *Handler) UpdateVideo(w http.ResponseWriter, r *http.Request) {
	var settings models.VideoSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateVideoSettings(&settings); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetCamera(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.GetCameraSettings()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(settings)
}

func (h *Handler) UpdateCamera(w http.ResponseWriter, r *http.Request) {
	var settings models.CameraSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateCameraSettings(&settings); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetTelemetry(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.GetTelemetrySettings()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(settings)
}

func (h *Handler) UpdateTelemetry(w http.ResponseWriter, r *http.Request) {
	var settings models.TelemetrySettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateTelemetrySettings(&settings); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetAdaptiveLink(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.GetAdaptiveLinkSettings()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(settings)
}

func (h *Handler) UpdateAdaptiveLink(w http.ResponseWriter, r *http.Request) {
	var settings models.AdaptiveLinkSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateAdaptiveLinkSettings(&settings); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetTxProfiles(w http.ResponseWriter, r *http.Request) {
	profiles, err := h.service.GetTxProfiles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(profiles)
}

func (h *Handler) UpdateTxProfiles(w http.ResponseWriter, r *http.Request) {
	var profiles []models.TxProfile
	if err := json.NewDecoder(r.Body).Decode(&profiles); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateTxProfiles(profiles); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
