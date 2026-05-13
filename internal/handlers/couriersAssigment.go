package handlers

import "net/http"

func (h *CourierHandler) CourierAssignmentHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.CourierHandler.CourierAssignmentHandler"
	if r.Method != http.MethodPost {
		h.log.Errorf("%s: Method not allowed", op)
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := extractUUIDFromPath(r.URL.Path, "/api/auto-assign/")
	if err != nil {
		h.log.Errorf("%s: Cannot extract UUID from path: %s", op, err)
		writeErrorResponse(w, http.StatusBadRequest, "Cannot extract UUID from path")
		return
	}

	courier, err := h.courierService.CourierAssignmentService(id)
	if err != nil {
		h.log.Errorf("%s: Cannot get courier: %s", op, err)
		writeErrorResponse(w, http.StatusInternalServerError, "Cannot get courier")
		return
	}

	writeJSONResponse(w, http.StatusOK, courier)
}
