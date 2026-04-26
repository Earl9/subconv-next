package config

import (
	"encoding/json"
	"fmt"

	"subconv-next/internal/model"
	"subconv-next/internal/storage"
)

func WriteJSON(path string, cfg model.Config) error {
	if err := Validate(cfg); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	data = append(data, '\n')

	if err := storage.AtomicWriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
