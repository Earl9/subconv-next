package nodestate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"subconv-next/internal/model"
	"subconv-next/internal/storage"
)

func Load(path string) (model.NodeState, error) {
	state := model.DefaultNodeState()
	path = strings.TrimSpace(path)
	if path == "" {
		return state, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil
		}
		return model.NodeState{}, fmt.Errorf("read state: %w", err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return state, nil
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&state); err != nil {
		return model.NodeState{}, fmt.Errorf("decode state: %w", err)
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return model.NodeState{}, fmt.Errorf("decode state: trailing content")
	}

	return model.NormalizeNodeState(state), nil
}

func Save(path string, state model.NodeState) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}

	state = model.NormalizeNodeState(state)
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	data = append(data, '\n')

	if err := storage.AtomicWriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	return nil
}
