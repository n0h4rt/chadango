package chadango

import (
	"errors"
	"regexp"
	"time"
)

const (
	WEBSOCKET_ORIGIN    = "http://st.chatango.com"
	PM_SERVER           = "ws://c1.chatango.com:8080/"
	EVENT_BUFFER_SIZE   = 30
	PING_INTERVAL       = 90 * time.Second
	DEFAULT_COLOR       = "000"
	DEFAULT_TEXT_FONT   = "1"
	DEFAULT_TEXT_SIZE   = 11
	MAX_MESSAGE_HISTORY = 100
	SYNC_SEND_TIMEOUT   = 5 * time.Second
	BASE_BACKOFF_DUR    = 1 * time.Second
	MAX_BACKOFF_DUR     = 30 * time.Second
	MAX_RETRIES         = 10
	MSG_LENGTH_DEFAULT  = 2900
	MSG_LENGTH_SHORT    = 850
	API_TIMEOUT         = 10 * time.Second
	BOT_DATA_KEY        = "<BOTDATA>"
)

const (
	_ int64 = 1 << iota
	_
	FlagPremium
	FlagBackground
	FlagMedia
	_
	FlagModIcon
	FlagStaffIcon
	FlagRedChannel
	_
	_
	FlagBlueChannel
	_
	_
	_
	FlagModChannel
)

const (
	charset = "abcdefghijklmnopqrstuvwxyz0123456789"
)

var (
	ErrNotAGroup        = errors.New("not a group")
	ErrNotConnected     = errors.New("not connected")
	ErrAlreadyConnected = errors.New("already connected")
	ErrConnectionClosed = errors.New("connection closed")
	ErrRetryEnds        = errors.New("connection closed")

	ErrLoginFailed     = errors.New("failed to login")
	ErrInvalidResponse = errors.New("invalid response")
	ErrNoArgument      = errors.New("no argument")
	ErrTimeout         = errors.New("timeout")

	ErrRateLimited          = errors.New("rate limited")
	ErrCLimited             = errors.New("climited")
	ErrMessageLength        = errors.New("message length exceeded")
	ErrFloodWarning         = errors.New("flood warning")
	ErrFloodBanned          = errors.New("flood banned")
	ErrNonSenseWarning      = errors.New("nonsense warning")
	ErrSpamWarning          = errors.New("spam warning")
	ErrShortWarning         = errors.New("short warning")
	ErrRestricted           = errors.New("restricted")
	ErrMustLogin            = errors.New("must login")
	ErrProxyBanned          = errors.New("proxy banned")
	ErrVerificationRequired = errors.New("verification required")

	ErrUpdateFailed    = errors.New("update failed")
	ErrClearFailed     = errors.New("clear failed")
	ErrAddModFailed    = errors.New("clear failed")
	ErrUpdateModFailed = errors.New("clear failed")
	ErrRemoveModFailed = errors.New("clear failed")

	ErrBadAlias = errors.New("bad alias")
	ErrBadLogin = errors.New("bad login")

	ErrPremiumExpired = errors.New("premium expired")

	ErrInvalidUsername = errors.New("invalid username")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrOfflineLimit    = errors.New("offline message limit")

	ErrSetTokenFailed = errors.New("settoken failed")
	ErrGCMRegFailed   = errors.New("GCM register failed")
	ErrGCMUnregFailed = errors.New("GCM unregister failed")
)

var (
	AnonSeedRe         = regexp.MustCompile(`<n\d{4}/>`)
	NameColorRe        = regexp.MustCompile(`<n([\da-fA-F]{1,6})\/>`)
	FontStyleRe        = regexp.MustCompile(`<f x([\da-fA-F]+)?="([\d\w]+)?">`)
	PrivateFontStyleRe = regexp.MustCompile(`<g x(\d+)?s([\da-fA-F]+)?="([\d\w]+)?">`)
	HtmlTagRe          = regexp.MustCompile(`<.*?>`)
	NameFontTag        = regexp.MustCompile(`<[nf][^>]*>`)
)

var GroupPermissions = map[string]int64{
	"DELETED":                1,
	"EDIT_MODS":              2,
	"EDIT_MOD_VISIBILITY":    4,
	"EDIT_BW":                8,
	"EDIT_RESTRICTIONS":      16,
	"EDIT_GROUP":             32,
	"SEE_COUNTER":            64,
	"SEE_MOD_CHANNEL":        128,
	"SEE_MOD_ACTIONS":        256,
	"EDIT_NLP":               512,
	"EDIT_GP_ANNC":           1024,
	"EDIT_ADMINS":            2048, // removed in current version
	"EDIT_SUPERMODS":         4096, // removed in current version
	"NO_SENDING_LIMITATIONS": 8192,
	"SEE_IPS":                16384,
	"CLOSE_GROUP":            32768,
	"CAN_BROADCAST":          65536,
	"MOD_ICON_VISIBLE":       131072,
	"IS_STAFF":               262144,
	"STAFF_ICON_VISIBLE":     524288,
	"UNBAN_ALL":              1048576,
}

var GroupStatuses = map[string]int64{
	"MISSING_1":                1,
	"NO_ANONS":                 4,
	"MISSING_2":                8,
	"NO_COUNTER":               16,
	"DISALLOW_IMAGES":          32,
	"DISALLOW_LINKS":           64,
	"DISALLOW_VIDEOS":          128,
	"MISSING_3":                256,
	"MISSING_4":                512,
	"BANWORD_ONLY_TO_AUTHOR":   1024,
	"FLOOD_CONTROLLED":         2048,
	"ENABLE_CHANNELS":          8192,
	"BASIC_NONSENSE_DETECTION": 16384, // js: nlp_single_msg
	"BLOCK_REPETITIOUS_MSGS":   32768, // js: nlp_msg_queue
	"BROADCAST_MODE":           65536,
	"CLOSED_NO_MODS":           131072,
	"GROUP_CLOSED":             262144,
	"DISPLAY_BADGES":           524288,
	"MODS_CHOOSE_BADGES":       1048576,
	"ADV_NONSENSE_DETECTION":   2097152, // js: nlp_ngram
	"BAN_PROXIES_AND_VPN":      4194304,
	"MISSING_5":                8388608,
	"MISSING_6":                268435456,
	"MISSING_7":                536870912,
}
