package entdefReader;

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	log "github.com/sirupsen/logrus"
)

func ReadEntdef(name string) (*EntityDefinition, error) {
	filename := "dat/entdefs/" + name + ".yml"

	log.Infof("Reading entdef %v", filename)

	file, err := ioutil.ReadFile(filename);

	if err != nil {
		log.Warnf("Cannot read entdef file: %v", err)
		return nil, err
	}

	entdef := EntityDefinition{}

	err = yaml.UnmarshalStrict(file, &entdef);

	if err != nil {
		log.Warnf("Cannot unmarshal entdef file: %v", err)
		return nil, err
	}

	return &entdef, err
}
