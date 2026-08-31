package updater

import (
	"strconv"

	"github.com/kamiertop/videodown/internal/constant"
)

func (u *Updater) SetAutoUpdate(enabled bool) error {
	return u.store.Set(constant.AutoUpdateKey, strconv.FormatBool(enabled))
}

func (u *Updater) IsAutoUpdate() (bool, error) {
	return u.store.AutoUpdate()
}
