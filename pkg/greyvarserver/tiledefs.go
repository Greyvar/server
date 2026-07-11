package greyvarserver

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/greyvar/datlib/common"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type tileDefYAML struct {
	Traversable *bool  `yaml:"traversable"`
	Texture     string `yaml:"texture"`
}

var blockingTileTextures map[string]bool

func loadBlockingTileTextures() map[string]bool {
	blocking := map[string]bool{
		// Editor defaults from dat/texdefs and defaultTextureAttributes.xml
		"water.png":   true,
		"barrier.png": true,
	}

	texdefsDir := filepath.Join(common.DatDir(), "texdefs", "tiles")
	entries, err := ioutil.ReadDir(texdefsDir)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Warnf("Could not read tiledefs directory %s: %v", texdefsDir, err)
		}
		return blocking
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		path := filepath.Join(texdefsDir, entry.Name())
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Warnf("Could not read tiledef %s: %v", path, err)
			continue
		}

		def := tileDefYAML{}
		if err := yaml.Unmarshal(data, &def); err != nil {
			log.Warnf("Could not parse tiledef %s: %v", path, err)
			continue
		}

		texture := def.Texture
		if texture == "" {
			texture = strings.TrimSuffix(entry.Name(), ".yaml") + ".png"
		}

		if def.Traversable != nil && !*def.Traversable {
			blocking[texture] = true
			log.Infof("Tiledef marks non-traversable: %s", texture)
		}
	}

	return blocking
}

func isBlockingTileTexture(texture string) bool {
	return blockingTileTextures[texture]
}
