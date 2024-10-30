package chadango

import (
	"errors"
	"time"
)

const (
	WEBSOCKET_ORIGIN    = "http://st.chatango.com"
	PM_SERVER           = "ws://c1.chatango.com:8080/"
	EVENT_BUFFER_SIZE   = 30
	PING_INTERVAL       = 90 * time.Second
	MAX_MESSAGE_HISTORY = 100
	SYNC_SEND_TIMEOUT   = 5 * time.Second
	BASE_BACKOFF_DUR    = 1 * time.Second
	MAX_BACKOFF_DUR     = 30 * time.Second
	MAX_RETRIES         = 10
	MSG_LENGTH_DEFAULT  = 2900
	MSG_LENGTH_SHORT    = 850
	API_TIMEOUT         = 10 * time.Second
)

const (
	API_LOGIN          = "https://chatango.com/login"
	API_GRP_LIST_UPD   = "https://chatango.com/groupslistupdate"
	API_MSG_BG_XML     = "https://ust.chatango.com/profileimg/%s/%s/%s/msgbg.xml"
	API_UPD_MSG_BG     = "https://chatango.com/updatemsgbg"
	API_SEARCH_PEOPLE  = "https://chatango.com/search"
	API_UPLOAD_IMG     = "https://chatango.com/uploadimg"
	API_MSG_STYLE_JSON = "https://ust.chatango.com/profileimg/%s/%s/%s/msgstyles.json"
	API_UPD_MSG_STYLE  = "https://chatango.com/updatemsgstyles"
	API_SET_TOKEN_GCM  = "https://chatango.com/settokenapp"
	API_REG_GCM        = "https://chatango.com/updategcm"
	API_UNREG_GCM      = "https://chatango.com/unregistergcm"

	API_CHECK_USER  = "https://st.chatango.com/script/namecheckeraccsales"
	API_CHECK_GROUP = "https://chatango.com/checkname"
	API_MINI_XML    = "https://ust.chatango.com/profileimg/%s/%s/%s/mod1.xml"
	API_FULL_XML    = "https://ust.chatango.com/profileimg/%s/%s/%s/mod2.xml"
)

var (
	ErrNotAGroup        = errors.New("not a group")
	ErrNotConnected     = errors.New("not connected")
	ErrAlreadyConnected = errors.New("already connected")
	ErrConnectionClosed = errors.New("connection closed")
	ErrRetryEnds        = errors.New("retry ends")
	ErrCLimited         = errors.New("climited")

	ErrLoginFailed = errors.New("failed to login")
	ErrNoArgument  = errors.New("no argument")
	ErrTimeout     = errors.New("timeout")

	ErrRateLimited          = errors.New("rate limited")
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
	ErrOfflineLimit         = errors.New("offline message limit")

	ErrBadAlias = errors.New("bad alias")
	ErrBadLogin = errors.New("bad login")

	ErrRequestFailed = errors.New("request failed")
)

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

type FontFamily int

const (
	FontFamilyArial FontFamily = iota
	FontFamilyComic
	FontFamilyGeorgia
	FontFamilyHandwriting
	FontFamilyImpact
	FontFamilyPalatino
	FontFamilyPapyrus
	FontFamilyTimes
	FontFamilyTypewriter
)
