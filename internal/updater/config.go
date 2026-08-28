package updater

func (u *Updater) SetAutoUpdate(enabled bool) error {
	return u.store.SetAutoUpdate(enabled)
}

func (u *Updater) IsAutoUpdate() (bool, error) {
	return u.store.AutoUpdate()
}
