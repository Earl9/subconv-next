package model

import "encoding/json"

func (c *SubscriptionConfig) UnmarshalJSON(data []byte) error {
	type alias SubscriptionConfig

	tmp := alias(DefaultSubscriptionConfig())
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*c = SubscriptionConfig(tmp)
	return nil
}

func (c *InlineConfig) UnmarshalJSON(data []byte) error {
	type alias InlineConfig

	tmp := alias(DefaultInlineConfig())
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*c = InlineConfig(tmp)
	return nil
}
