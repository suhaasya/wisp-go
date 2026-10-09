package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/suhaasya/wisp-go/internal/store"
)

const connectionsPrefix = "/api/v1/connections"

// ConnectionsHandler serves connection profile CRUD.
type ConnectionsHandler struct {
	profiles *store.ProfileStore
	secrets  store.Secrets
}

// NewConnectionsHandler returns a connections API handler.
func NewConnectionsHandler(profiles *store.ProfileStore, secrets store.Secrets) *ConnectionsHandler {
	return &ConnectionsHandler{profiles: profiles, secrets: secrets}
}

type connectionView struct {
	store.Profile
	HasPassword bool `json:"has_password"`
}

type connectionsListResponse struct {
	Version     int              `json:"version"`
	Folders     []string         `json:"folders"`
	Connections []connectionView `json:"connections"`
	Warning     string           `json:"warning,omitempty"`
}

// ServeHTTP implements http.Handler.
func (h *ConnectionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, connectionsPrefix)
	path = strings.TrimPrefix(path, "/")
	switch {
	case path == "" && r.Method == http.MethodGet:
		h.list(w, r)
	case path == "" && r.Method == http.MethodPost:
		h.create(w, r)
	case strings.HasSuffix(path, "/duplicate") && r.Method == http.MethodPost:
		id := strings.TrimSuffix(path, "/duplicate")
		id = strings.TrimSuffix(id, "/")
		h.duplicate(w, r, id)
	case path != "" && r.Method == http.MethodPut:
		h.update(w, r, path)
	case path != "" && r.Method == http.MethodDelete:
		h.delete(w, r, path)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (h *ConnectionsHandler) list(w http.ResponseWriter, _ *http.Request) {
	doc := h.profiles.Document()
	out := connectionsListResponse{
		Version: doc.Version,
		Folders: doc.Folders,
		Warning: doc.Warning,
	}
	if out.Folders == nil {
		out.Folders = []string{}
	}
	out.Connections = make([]connectionView, 0, len(doc.Connections))
	for _, p := range doc.Connections {
		out.Connections = append(out.Connections, h.viewFor(p))
	}
	writeJSON(w, out)
}

func (h *ConnectionsHandler) viewFor(p store.Profile) connectionView {
	has, _ := h.hasDBPassword(p.ID)
	return connectionView{Profile: p, HasPassword: has}
}

func (h *ConnectionsHandler) hasDBPassword(profileID string) (bool, error) {
	if h.secrets == nil {
		return false, nil
	}
	val, ok, err := h.secrets.Get(profileID, store.SecretKindDBPassword)
	return ok && len(val) > 0, err
}

func (h *ConnectionsHandler) create(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "could not read body", "")
		return
	}
	profile, secretFields, err := decodeProfileWrite(body, "")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error(), "")
		return
	}
	created, err := h.profiles.Create(profile)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	if err := h.applySecretFields(created.ID, secretFields); err != nil {
		_ = h.profiles.Delete(created.ID)
		writeError(w, http.StatusInternalServerError, "secrets_write", "could not store secrets", err.Error())
		return
	}
	writeJSON(w, h.viewFor(created))
}

func (h *ConnectionsHandler) update(w http.ResponseWriter, r *http.Request, id string) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "could not read body", "")
		return
	}
	if _, ok := h.profiles.Get(id); !ok {
		writeError(w, http.StatusNotFound, "not_found", "profile not found", "")
		return
	}
	profile, secretFields, err := decodeProfileWrite(body, id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error(), "")
		return
	}
	if err := h.profiles.Update(profile); err != nil {
		writeProfileError(w, err)
		return
	}
	if err := h.applySecretUpdates(id, secretFields); err != nil {
		writeError(w, http.StatusInternalServerError, "secrets_write", "could not update secrets", err.Error())
		return
	}
	updated, _ := h.profiles.Get(id)
	writeJSON(w, h.viewFor(updated))
}

func (h *ConnectionsHandler) delete(w http.ResponseWriter, _ *http.Request, id string) {
	if err := h.profiles.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, "not_found", "profile not found", "")
		return
	}
	if h.secrets != nil {
		_ = h.secrets.DeleteAll(id)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ConnectionsHandler) duplicate(w http.ResponseWriter, _ *http.Request, id string) {
	dup, err := h.profiles.Duplicate(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "profile not found", "")
		return
	}
	if h.secrets != nil {
		for _, kind := range []store.SecretKind{
			store.SecretKindDBPassword,
			store.SecretKindSSHPassword,
			store.SecretKindSSHPassphrase,
		} {
			val, ok, err := h.secrets.Get(id, kind)
			if err != nil || !ok {
				continue
			}
			if err := h.secrets.Set(dup.ID, kind, val); err != nil {
				_ = h.profiles.Delete(dup.ID)
				writeError(w, http.StatusInternalServerError, "secrets_write", "could not copy secrets", err.Error())
				return
			}
		}
	}
	writeJSON(w, h.viewFor(dup))
}

type secretWriteFields struct {
	password      *secretField
	sshPassword   *secretField
	sshPassphrase *secretField
}

type secretField struct {
	present bool
	value   []byte
}

func decodeProfileWrite(body []byte, id string) (store.Profile, secretWriteFields, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return store.Profile{}, secretWriteFields{}, err
	}
	secrets := secretWriteFields{}
	if v, ok := raw["password"]; ok {
		secrets.password = &secretField{present: true}
		delete(raw, "password")
		if string(v) != "null" {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				return store.Profile{}, secretWriteFields{}, err
			}
			secrets.password.value = []byte(s)
		}
	}
	if v, ok := raw["ssh_password"]; ok {
		secrets.sshPassword = &secretField{present: true}
		delete(raw, "ssh_password")
		if string(v) != "null" {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				return store.Profile{}, secretWriteFields{}, err
			}
			secrets.sshPassword.value = []byte(s)
		}
	}
	if v, ok := raw["ssh_passphrase"]; ok {
		secrets.sshPassphrase = &secretField{present: true}
		delete(raw, "ssh_passphrase")
		if string(v) != "null" {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				return store.Profile{}, secretWriteFields{}, err
			}
			secrets.sshPassphrase.value = []byte(s)
		}
	}
	delete(raw, "id")
	delete(raw, "has_password")
	rest, err := json.Marshal(raw)
	if err != nil {
		return store.Profile{}, secretWriteFields{}, err
	}
	var p store.Profile
	if err := json.Unmarshal(rest, &p); err != nil {
		return store.Profile{}, secretWriteFields{}, err
	}
	if id != "" {
		p.ID = id
	}
	return p, secrets, nil
}

func (h *ConnectionsHandler) applySecretFields(id string, f secretWriteFields) error {
	if h.secrets == nil {
		return nil
	}
	if f.password != nil && len(f.password.value) > 0 {
		if err := h.secrets.Set(id, store.SecretKindDBPassword, f.password.value); err != nil {
			return err
		}
	}
	if f.sshPassword != nil && len(f.sshPassword.value) > 0 {
		if err := h.secrets.Set(id, store.SecretKindSSHPassword, f.sshPassword.value); err != nil {
			return err
		}
	}
	if f.sshPassphrase != nil && len(f.sshPassphrase.value) > 0 {
		if err := h.secrets.Set(id, store.SecretKindSSHPassphrase, f.sshPassphrase.value); err != nil {
			return err
		}
	}
	return nil
}

func (h *ConnectionsHandler) applySecretUpdates(id string, f secretWriteFields) error {
	if h.secrets == nil {
		return nil
	}
	if f.password != nil {
		if len(f.password.value) == 0 {
			if err := deleteSecretKind(h.secrets, id, store.SecretKindDBPassword); err != nil {
				return err
			}
		} else {
			if err := h.secrets.Set(id, store.SecretKindDBPassword, f.password.value); err != nil {
				return err
			}
		}
	}
	if f.sshPassword != nil {
		if len(f.sshPassword.value) == 0 {
			if err := deleteSecretKind(h.secrets, id, store.SecretKindSSHPassword); err != nil {
				return err
			}
		} else if err := h.secrets.Set(id, store.SecretKindSSHPassword, f.sshPassword.value); err != nil {
			return err
		}
	}
	if f.sshPassphrase != nil {
		if len(f.sshPassphrase.value) == 0 {
			if err := deleteSecretKind(h.secrets, id, store.SecretKindSSHPassphrase); err != nil {
				return err
			}
		} else if err := h.secrets.Set(id, store.SecretKindSSHPassphrase, f.sshPassphrase.value); err != nil {
			return err
		}
	}
	return nil
}

func deleteSecretKind(sec store.Secrets, profileID string, kind store.SecretKind) error {
	type deleter interface {
		Delete(profileID string, kind store.SecretKind) error
	}
	if d, ok := sec.(deleter); ok {
		return d.Delete(profileID, kind)
	}
	return sec.Set(profileID, kind, nil)
}

func writeProfileError(w http.ResponseWriter, err error) {
	if err.Error() == "profile not found" {
		writeError(w, http.StatusNotFound, "not_found", err.Error(), "")
		return
	}
	msg := err.Error()
	writeError(w, http.StatusBadRequest, "invalid_profile", msg, "")
}
