package main

import "net/http"

func (cfg apiConfig) resetDatabase(w http.ResponseWriter, r *http.Request) {

	err := cfg.db.DeleteUsers(r.Context())

	if err != nil {
		respondWithError(w, http.StatusNotModified, "Failed to delete all users and their data", err, false)
		return
	}

	respondWithJSON(w, http.StatusOK, nil)
}
