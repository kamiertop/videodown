package storage

// CloseAction records the user's choice for closing the main window.
type CloseAction uint8

const (
	CloseActionUnset CloseAction = iota
	CloseActionExit
	CloseActionTray
)

func closeAction(toTray bool) CloseAction {
	if toTray {
		return CloseActionTray
	}
	return CloseActionExit
}

func (s *Store) CloseToTray() (enabled, configured bool, err error) {
	action := s.Config().App.CloseAction
	return action != CloseActionExit, action != CloseActionUnset, nil
}

func (s *Store) SetCloseToTray(enabled bool) error {
	return s.UpdateConfig(func(c *Config) error {
		c.App.CloseAction = closeAction(enabled)
		return nil
	})
}
