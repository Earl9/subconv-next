package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"subconv-next/internal/model"
)

func LoadJSONBytes(data []byte) (model.Config, error) {
	cfg := model.DefaultConfig()

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return model.Config{}, fmt.Errorf("decode JSON config: %w", err)
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return model.Config{}, fmt.Errorf("decode JSON config: trailing content")
	}

	if err := validateConfig(cfg); err != nil {
		return model.Config{}, err
	}

	return cfg, nil
}
