package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	db "TP_WEB/db/sqlc"

	"TP_WEB/service"
)

type UsuarioHandler struct {
	service  *service.UsuarioService
	formPath string
}

func NewUsuarioHandler(usuarioService *service.UsuarioService, formPath string) *UsuarioHandler {
	return &UsuarioHandler{service: usuarioService, formPath: formPath}
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
		slog.WarnContext(r.Context(), "alta de usuario rechazada", "motivo", "datos_invalidos")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	usuario, err := h.service.Crear(r.Context(), email, contrasena)
	if errors.Is(err, service.ErrEmailInvalido) || errors.Is(err, service.ErrContrasenaCorta) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, service.ErrEmailExistente) {
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
	_ = json.NewEncoder(w).Encode(respuestaUsuario(usuario))
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

// La respuesta pública del usuario excluye la contraseña en todas las operaciones.
type usuarioRespuesta struct {
	ID    int32  `json:"id_usuario"`
	Email string `json:"email"`
}

func respuestaUsuario(usuario db.Usuario) usuarioRespuesta {
	return usuarioRespuesta{ID: usuario.IDUsuario, Email: usuario.Email}
}

// POST /usuarios y GET /usuarios.
func (h *UsuarioHandler) Coleccion(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.crearUsuario(w, r)
	case http.MethodGet:
		h.listar(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
	}
}

// GET, PUT y DELETE /usuarios/{id}.
func (h *UsuarioHandler) Elemento(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
	if err != nil || id <= 0 {
		http.Error(w, "ID de usuario invalido", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.obtenerPorID(w, r, int32(id))
	case http.MethodPut:
		h.actualizar(w, r, int32(id))
	case http.MethodDelete:
		h.eliminar(w, r, int32(id))
	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func (h *UsuarioHandler) listar(w http.ResponseWriter, r *http.Request) {
	usuarios, err := h.service.Listar(r.Context())
	if err != nil {
		http.Error(w, "error al obtener usuarios", http.StatusInternalServerError)
		return
	}
	respuesta := make([]usuarioRespuesta, 0, len(usuarios))
	for _, usuario := range usuarios {
		respuesta = append(respuesta, respuestaUsuario(usuario))
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(respuesta)
}

func (h *UsuarioHandler) obtenerPorID(w http.ResponseWriter, r *http.Request, id int32) {
	usuario, err := h.service.ObtenerPorID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUsuarioNoEncontrado) {
			http.Error(w, "usuario no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "error al obtener el usuario", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(respuestaUsuario(usuario))
}

func (h *UsuarioHandler) actualizar(w http.ResponseWriter, r *http.Request, id int32) {
	var dto service.ActualizarUsuarioDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	usuario, err := h.service.Actualizar(r.Context(), id, dto)
	if err != nil {
		if errors.Is(err, service.ErrUsuarioNoEncontrado) {
			http.Error(w, "usuario no encontrado", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrEmailInvalido) || errors.Is(err, service.ErrContrasenaCorta) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrEmailExistente) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "error al actualizar el usuario", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(respuestaUsuario(usuario))
}

func (h *UsuarioHandler) eliminar(w http.ResponseWriter, r *http.Request, id int32) {
	if err := h.service.Eliminar(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrUsuarioNoEncontrado) {
			http.Error(w, "usuario no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "error al eliminar el usuario", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
