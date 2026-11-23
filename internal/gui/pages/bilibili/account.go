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

// authState 汇总登录与会话状态：扫码二维码、轮询、账号资料与登录错误。
type authState struct {
	loggedIn, checking, loggingOut bool
	error                          string
	account                        model.MyInfoProfile
	pollSeq                        uint64
	cancelPoll                     func()
	qr                             resource[model.QRCodeData]
	qrImage                        *ui.Bitmap
	qrExpired                      bool
	qrStatus                       string
}

func (s *State) checkAccount() {
	if s.auth.checking || s.api == nil {
		return
	}
	s.auth.checking = true
	s.auth.error = ""
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
		s.auth.checking = false
		s.auth.loggedIn = loggedIn
		if err != nil {
			s.auth.error = err.Error()
		} else {
			s.auth.account = profile
		}
		if refreshError != nil && s.auth.error == "" {
			s.auth.error = "刷新登录凭证失败：" + refreshError.Error()
		}
		s.enter(s.section)
		if !loggedIn && s.section == "profile" {
			s.newQRCode()
		}
	})
}

func (s *State) stopPoll() {
	s.auth.pollSeq++
	if s.auth.cancelPoll != nil {
		s.auth.cancelPoll()
		s.auth.cancelPoll = nil
	}
}
func (s *State) newQRCode() {
	if s.api == nil || s.auth.qr.Loading || s.auth.checking || s.auth.loggingOut || s.auth.loggedIn {
		return
	}
	s.stopPoll()
	s.auth.qrImage = nil
	s.auth.qrExpired = false
	s.auth.qrStatus = "正在获取二维码"
	load(s, &s.auth.qr, func() (model.QRCodeData, error) {
		data, err := s.api.QRCode()
		if err == nil && (strings.TrimSpace(data.Url) == "" || strings.TrimSpace(data.QRCodeKey) == "") {
			err = errors.New("二维码响应为空")
		}
		return data, err
	}, func(data model.QRCodeData) {
		qr, err := qrcode.New(data.Url, qrcode.Medium)
		if err != nil {
			s.auth.qr.Error = err.Error()
			return
		}
		s.auth.qrImage = ui.NewBitmap(qr.Image(int(style.QRCodeSize))).Pixelated()
		s.auth.qrStatus = "请使用 B 站 App 扫码"
		if s.section == "profile" {
			s.schedulePoll()
		}
	})
}
func (s *State) schedulePoll() {
	if s.runtime.After == nil || s.stopped || s.section != "profile" || s.auth.loggedIn || s.auth.qrExpired {
		return
	}
	s.stopPoll()
	seq := s.auth.pollSeq
	s.auth.cancelPoll = s.runtime.After(1200*time.Millisecond, func() {
		s.auth.cancelPoll = nil
		if seq != s.auth.pollSeq || s.stopped || s.section != "profile" {
			return
		}
		s.pollQRCode(seq)
	})
}
func (s *State) pollQRCode(seq uint64) {
	key, epoch := s.auth.qr.Data.QRCodeKey, s.epoch
	if key == "" {
		return
	}
	var data model.PollQRCodeData
	s.runtime.Run(func() error { var err error; data, err = s.api.PollQRCode(key); return err }, func(err error) {
		if s.stopped || epoch != s.epoch || seq != s.auth.pollSeq {
			return
		}
		if err != nil {
			s.auth.qr.Error = err.Error()
			return
		}
		switch data.Code {
		case model.PollQRCodeStatusSuccess:
			s.stopPoll()
			s.auth.loggedIn = true
			s.auth.qrImage = nil
			s.auth.qr = resource[model.QRCodeData]{}
			s.checkAccount()
		case model.PollQRCodeStatusExpired:
			s.auth.qrExpired = true
			s.auth.qrStatus = "二维码已过期"
		case model.PollQRCodeStatusScannedUnconfirmed:
			s.auth.qrStatus = "已扫码，请在手机上确认"
			s.schedulePoll()
		case model.PollQRCodeStatusNotScanned:
			s.auth.qrStatus = "请使用 B 站 App 扫码"
			s.schedulePoll()
		default:
			s.auth.qr.Error = "未知二维码状态，请重新获取"
		}
	})
}

func (s *State) logout() {
	if s.api == nil || s.auth.loggingOut || s.auth.checking {
		return
	}
	s.stopPoll()
	for _, b := range s.batches.sessions {
		s.stopBatch(b)
	}
	s.auth.loggingOut = true
	s.auth.error = ""
	epoch := s.epoch
	var remains bool
	s.runtime.Run(func() error { _, err := s.api.LogOut(); remains = s.api.IsLoggedIn(); return err }, func(err error) {
		if s.stopped || epoch != s.epoch {
			return
		}
		s.auth.loggingOut = false
		if !remains {
			s.epoch++
			s.auth.loggedIn = false
			s.auth.account = model.MyInfoProfile{}
			s.clearPrivateData()
			s.auth.qr = resource[model.QRCodeData]{}
			s.auth.qrImage = nil
			s.auth.qrExpired = false
		}
		if err != nil {
			s.auth.error = err.Error()
		}
		if !remains && s.section == "profile" {
			s.newQRCode()
		}
	})
}
