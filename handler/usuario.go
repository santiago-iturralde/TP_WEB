package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"TP_WEB/repository"
)

type UsuarioHandler struct {
	repository repository.UsuarioRepository
	formPath   string
}

func NewUsuarioHandler(repo repository.UsuarioRepository, formPath string) *UsuarioHandler {
	return &UsuarioHandler{repository: repo, formPath: formPath}
}

func (h *UsuarioHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.mostrarFormulario(w, r)
	case http.MethodPost:
		h.crearUsuario(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func (h *UsuarioHandler) mostrarFormulario(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, h.formPath)
}

func (h *UsuarioHandler) crearUsuario(w http.ResponseWriter, r *http.Request) {
	email, contrasena, err := datosUsuario(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	email = strings.ToLower(strings.TrimSpace(email))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		http.Error(w, "el email no es valido", http.StatusBadRequest)
		return
	}
	if len(contrasena) < 8 {
		http.Error(w, "la contrasena debe tener al menos 8 caracteres", http.StatusBadRequest)
		return
	}
	usuario, err := h.repository.Crear(r.Context(), email, contrasena)
	if errors.Is(err, repository.ErrEmailExistente) {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "no se pudo crear el usuario", http.StatusInternalServerError)
		return
	}

	// La contraseña no se incluye en la respuesta.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(struct {
		ID    int32  `json:"id_usuario"`
		Email string `json:"email"`
	}{ID: usuario.IDUsuario, Email: usuario.Email})
}

func datosUsuario(r *http.Request) (string, string, error) {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		var datos struct {
			Email      string `json:"email"`
			Contrasena string `json:"contrasena"`
		}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&datos); err != nil {
			return "", "", errors.New("JSON invalido")
		}
		return datos.Email, datos.Contrasena, nil
	}

	if err := r.ParseForm(); err != nil {
		return "", "", errors.New("formulario invalido")
	}
	return r.FormValue("email"), r.FormValue("contrasena"), nil
}
