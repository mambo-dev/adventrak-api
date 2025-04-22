package main

import (
	"net/http"
	"os"
)

func (cfg apiConfig) resetDatabase(w http.ResponseWriter, r *http.Request) {

	dirs, err := os.ReadDir(cfg.assetsRoot)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err, false)
		return
	}

	for _, dir := range dirs {
		if err := os.Remove(dir.Name()); err != nil {
			respondWithError(w, http.StatusNotModified, "Failed to delete user assets and the assets directory.", err, false)
			return
		}
	}

	err = cfg.db.DeleteUsers(r.Context())

	if err != nil {
		respondWithError(w, http.StatusNotModified, "Failed to delete all users and their data", err, false)
		return
	}

	respondWithJSON(w, http.StatusOK, nil)
}
