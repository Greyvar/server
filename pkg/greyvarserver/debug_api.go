package greyvarserver

import (
	"encoding/json"
	"net/http"
	"sort"
)

func (s *serverInterface) handleBlockingTiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	textures := make([]string, 0, len(blockingTileTextures))
	for texture := range blockingTileTextures {
		textures = append(textures, texture)
	}

	sort.Strings(textures)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string][]string{
		"textures": textures,
	})
}
