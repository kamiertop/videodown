package storage

func (s *Store) CloseToTray() (enabled, configured bool, err error) {
	choice := s.Config().App.CloseToTray
	if choice == nil {
		return true, false, nil
	}
	return *choice, true, nil
}

func (s *Store) SetCloseToTray(enabled bool) error {
	return s.UpdateConfig(func(c *Config) error { c.App.CloseToTray = &enabled; return nil })
}
