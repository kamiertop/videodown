package model

const (
	SuccessCode = 0
)

const (
	PollQRCodeStatusSuccess            uint32 = 0
	PollQRCodeStatusNotScanned         uint32 = 86101
	PollQRCodeStatusExpired            uint32 = 86038
	PollQRCodeStatusScannedUnconfirmed uint32 = 86090
)
