package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"subconv-next/internal/model"
)

func LoadJSONBytes(data []byte) (model.Config, error) {
	return loadJSONBytes(data, false)
}

// LoadJSONBytesForEditing allows stored custom rules that predate validation to
// remain editable. Other config fields are still validated; callers must use
// the strict validator before saving or generating a publication.
func LoadJSONBytesForEditing(data []byte) (model.Config, error) {
	return loadJSONBytes(data, true)
}

func loadJSONBytes(data []byte, forEditing bool) (model.Config, error) {
	cfg := model.DefaultConfig()

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return model.Config{}, fmt.Errorf("decode JSON config: %w", err)
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return model.Config{}, fmt.Errorf("decode JSON config: trailing content")
	}

	cfg = normalizeConfig(cfg)

	validationConfig := cfg
	if forEditing {
		validationConfig.Render.CustomRules = nil
		validationConfig.Render.AdditionalRules = nil
	}
	if err := validateConfig(validationConfig); err != nil {
		return model.Config{}, err
	}

	return cfg, nil
}
