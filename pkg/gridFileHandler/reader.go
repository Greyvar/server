package gridFileHandler;

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

func ReadGridFile(filename string) (*GridFile, error) {
	file, err := ioutil.ReadFile(filename);

	if err != nil {
		return nil, err
	}

	gf := GridFile{}

	err = yaml.UnmarshalStrict(file, &gf);

	if err != nil {
		return nil, err
	}

	return &gf, err
}
