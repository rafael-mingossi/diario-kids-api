package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	authmiddleware "github.com/rafael-mingossi/diario-kids-api/internal/middleware"
	"github.com/rafael-mingossi/diario-kids-api/internal/services"
)

func registrarAuditoriaBestEffort(service services.AuditService, r *http.Request, input services.AuditLogInput) {
	if service == nil {
		return
	}

	input.IP = r.RemoteAddr
	input.UserAgent = r.UserAgent()

	if err := service.Registrar(input); err != nil {
		slog.Error("falha ao registrar auditoria",
			"path", r.URL.Path,
			"detalhe", err,
		)
	}
}

func actorFromRequest(r *http.Request) (*uint, string, string, *uint) {
	usuario, ok := r.Context().Value(authmiddleware.ContextKeyUsuario).(authmiddleware.UsuarioAutenticado)
	if !ok {
		return nil, "", "", nil
	}

	actorID64, err := strconv.ParseUint(usuario.ID, 10, 64)
	if err != nil {
		return nil, usuario.PlatformRole, usuario.Role, usuario.EscolaID
	}

	actorID := uint(actorID64)
	return &actorID, usuario.PlatformRole, usuario.Role, usuario.EscolaID
}
