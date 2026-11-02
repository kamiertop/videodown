package bilibili

import (
	"errors"
	"strings"
	"time"

	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/style"

	"github.com/egoist/mygo/ui"
	"github.com/skip2/go-qrcode"
)

func (s *State) checkAccount() {
	if s.checking || s.api == nil {
		return
	}
	s.checking = true
	s.authError = ""
	epoch := s.epoch
	var loggedIn bool
	var profile model.MyInfoProfile
	var refreshError error
	s.runtime.Run(func() error {
		loggedIn = s.api.IsLoggedIn()
		if !loggedIn {
			return nil
		}
		_, refreshError = s.api.RefreshCookie()
		var err error
		profile, err = s.api.MyInfo()
		return err
	}, func(err error) {
		if s.stopped || epoch != s.epoch {
			return
		}
		s.checking = false
		s.loggedIn = loggedIn
		if err != nil {
			s.authError = err.Error()
		} else {
			s.account = profile
		}
		if refreshError != nil && s.authError == "" {
			s.authError = "刷新登录凭证失败：" + refreshError.Error()
		}
		s.enter(s.section)
		if !loggedIn && s.section == "profile" {
			s.newQRCode()
		}
	})
}

func (s *State) stopPoll() {
	s.pollSeq++
	if s.cancelPoll != nil {
		s.cancelPoll()
		s.cancelPoll = nil
	}
}
func (s *State) newQRCode() {
	if s.api == nil || s.qr.Loading || s.checking || s.loggingOut || s.loggedIn {
		return
	}
	s.stopPoll()
	s.qrImage = nil
	s.qrExpired = false
	s.qrStatus = "正在获取二维码"
	load(s, &s.qr, func() (model.QRCodeData, error) {
		data, err := s.api.QRCode()
		if err == nil && (strings.TrimSpace(data.Url) == "" || strings.TrimSpace(data.QRCodeKey) == "") {
			err = errors.New("二维码响应为空")
		}
		return data, err
	}, func(data model.QRCodeData) {
		qr, err := qrcode.New(data.Url, qrcode.Medium)
		if err != nil {
			s.qr.Error = err.Error()
			return
		}
		s.qrImage = ui.NewBitmap(qr.Image(int(style.QRCodeSize))).Pixelated()
		s.qrStatus = "请使用 B 站 App 扫码"
		if s.section == "profile" {
			s.schedulePoll()
		}
	})
}
func (s *State) schedulePoll() {
	if s.runtime.After == nil || s.stopped || s.section != "profile" || s.loggedIn || s.qrExpired {
		return
	}
	s.stopPoll()
	seq := s.pollSeq
	s.cancelPoll = s.runtime.After(1200*time.Millisecond, func() {
		s.cancelPoll = nil
		if seq != s.pollSeq || s.stopped || s.section != "profile" {
			return
		}
		s.pollQRCode(seq)
	})
}
func (s *State) pollQRCode(seq uint64) {
	key, epoch := s.qr.Data.QRCodeKey, s.epoch
	if key == "" {
		return
	}
	var data model.PollQRCodeData
	s.runtime.Run(func() error { var err error; data, err = s.api.PollQRCode(key); return err }, func(err error) {
		if s.stopped || epoch != s.epoch || seq != s.pollSeq {
			return
		}
		if err != nil {
			s.qr.Error = err.Error()
			return
		}
		switch data.Code {
		case model.PollQRCodeStatusSuccess:
			s.stopPoll()
			s.loggedIn = true
			s.qrImage = nil
			s.qr = resource[model.QRCodeData]{}
			s.checkAccount()
		case model.PollQRCodeStatusExpired:
			s.qrExpired = true
			s.qrStatus = "二维码已过期"
		case model.PollQRCodeStatusScannedUnconfirmed:
			s.qrStatus = "已扫码，请在手机上确认"
			s.schedulePoll()
		case model.PollQRCodeStatusNotScanned:
			s.qrStatus = "请使用 B 站 App 扫码"
			s.schedulePoll()
		default:
			s.qr.Error = "未知二维码状态，请重新获取"
		}
	})
}

func (s *State) logout() {
	if s.api == nil || s.loggingOut || s.checking {
		return
	}
	s.stopPoll()
	for _, b := range s.batches {
		s.stopBatch(b)
	}
	s.loggingOut = true
	s.authError = ""
	epoch := s.epoch
	var remains bool
	s.runtime.Run(func() error { _, err := s.api.LogOut(); remains = s.api.IsLoggedIn(); return err }, func(err error) {
		if s.stopped || epoch != s.epoch {
			return
		}
		s.loggingOut = false
		if !remains {
			s.epoch++
			s.loggedIn = false
			s.account = model.MyInfoProfile{}
			s.clearPrivateData()
			s.qr = resource[model.QRCodeData]{}
			s.qrImage = nil
			s.qrExpired = false
		}
		if err != nil {
			s.authError = err.Error()
		}
		if !remains && s.section == "profile" {
			s.newQRCode()
		}
	})
}
