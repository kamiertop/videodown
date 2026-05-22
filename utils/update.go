package utils

const autoUpdateKey = "auto_update"

func (s *Settings) IsAutoUpdate() (bool, error) {
	key, err := s.GetKey(autoUpdateKey)
	if err != nil {
		return false, err
	}
	if key == "true" {
		return true, nil
	}

	return false, nil
}

func (s *Settings) SetAutoUpdate(enable bool) error {
	var value string
	if enable {
		value = "true"
	} else {
		value = "false"
	}

	return s.SetKey(autoUpdateKey, value)
}
