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
)

const (
	API_LOGIN          = "https://chatango.com/login"
	API_GRP_LIST_UPD   = "https://chatango.com/groupslistupdate"
	API_MSG_BG_IMG     = "https://ust.chatango.com/profileimg/%s/%s/%s/msgbg.jpg"
	API_MSG_BG_XML     = "https://ust.chatango.com/profileimg/%s/%s/%s/msgbg.xml"
	API_UPD_MSG_BG     = "https://chatango.com/updatemsgbg"
	API_SEARCH_PEOPLE  = "https://chatango.com/search"
	API_UPLOAD_IMG     = "https://chatango.com/uploadimg"
	API_MSG_STYLE_JSON = "https://ust.chatango.com/profileimg/%s/%s/%s/msgstyles.json"
	API_UPD_MSG_STYLE  = "https://chatango.com/updatemsgstyles"
	API_SET_TOKEN_GCM  = "https://chatango.com/settokenapp"
	API_REG_GCM        = "https://chatango.com/updategcm"
	API_UNREG_GCM      = "https://chatango.com/unregistergcm"
	API_UM_THUMB       = "https://ust.chatango.com/um/%s/%s/%s/img/t_%%d.jpg"
	API_UM_LARGE       = "https://ust.chatango.com/um/%s/%s/%s/img/l_%%d.jpg"

	API_CHECK_USER     = "https://st.chatango.com/script/namecheckeraccsales"
	API_CHECK_GROUP    = "https://chatango.com/checkname"
	API_MINI_XML       = "https://ust.chatango.com/profileimg/%s/%s/%s/mod1.xml"
	API_FULL_XML       = "https://ust.chatango.com/profileimg/%s/%s/%s/mod2.xml"
	API_PHOTO_FULL_IMG = "https://fp.chatango.com/profileimg/%s/%s/%s/full.jpg"
)

const (
	_ int64 = 1 << iota
	_
	FlagPremium
	FlagBackground
	FlagMedia
	FlagCensored
	FlagModIcon
	FlagStaffIcon
	FlagRedChannel  // js: #ed1c24
	_               // js: #ee7f22
	_               // js: #39b54a
	FlagBlueChannel // js: #25aae1
	_               // js: #0e76bc
	_               // js: #662d91
	_               // js: #ed217c
	FlagModChannel
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

var (
	AnonSeedRe         = regexp.MustCompile(`<n\d{4}/>`)
	NameColorRe        = regexp.MustCompile(`<n([\da-fA-F]{1,6})\/>`)
	FontStyleRe        = regexp.MustCompile(`<f x([\da-fA-F]+)?="([\d\w]+)?">`)
	NameFontTag        = regexp.MustCompile(`<[nf]\s[^>]*>`)
	PrivateFontStyleRe = regexp.MustCompile(`<g x(\d+)?s([\da-fA-F]+)?="([\d\w]+)?">`)

	// Go does not support negative lookahead `<(?!br\s*\/?>).*?>`.
	// This alternative will match either the `<br>` and `<br/>` tags (captured in group 1)
	// or any other HTML tags (captured in group 2).
	// Then the `ReplaceAllString(text, "$1")` method will then keep the content matched by group 1
	// and remove the content matched by group 2.
	HtmlTagRe = regexp.MustCompile(`(<br\s*\/?>)|(<[^>]+>)`)
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
