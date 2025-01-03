package models

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/n0h4rt/chadango/utils"
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

var ModactionTmpl = map[string]string{
	"mod_actions_title":           "Moderator Actions Log",
	"mod_actions_title_sm":        "Moderator Log",
	"action_desc_amod":            "*name**ip* made *target* a moderator",
	"action_desc_aadm":            "*name**ip* made *target* an administrator",
	"action_desc_rmod":            "*name**ip* removed *target* as a moderator",
	"action_desc_radm":            "*name**ip* removed *target* as an administrator",
	"action_desc_emod":            "*name**ip*",
	"added_perm":                  "gave *target* permission to: *added*",
	"and":                         "and",
	"removed_perm":                "removed *target*'s permission to: *removed*",
	"action_desc_chrl":            "*name**ip* changed rate limit to *rl*",
	"action_desc_anon":            "*name**ip* *didallow* anons in the group",
	"action_desc_prxy":            "*name**ip* *didallow* messaging from proxies and VPNs in the group",
	"action_desc_brdc":            "*name**ip* *didenable* broadcast mode",
	"action_desc_cinm":            "*name**ip* *didenable* closed without moderators mode",
	"action_desc_acls":            "Group auto-closed since no moderators were online",
	"action_desc_aopn":            "Group re-opened upon moderator login",
	"action_desc_chan":            "*name**ip* *didenable* channels in the group",
	"action_desc_cntr":            "*name**ip* *didenable* the counter in the group",
	"action_desc_chbw":            "*name**ip* changed banned words",
	"action_desc_enlp":            "*name**ip* changed auto-moderation to",
	"action_desc_annc":            "*name**ip* *didenable* auto-announcement",
	"action_desc_egrp":            "*name**ip* edited group",
	"action_desc_shwi":            "*name**ip* forced moderators to show a badge",
	"action_desc_hidi":            "*name**ip* hid all moderator badges",
	"action_desc_chsi":            "*name**ip* let moderators choose their own badge visibility",
	"action_desc_ubna":            "*name**ip* removed all bans",
	"enable_annc":                 "every *n* seconds: *msg*",
	"perm_edit_mods":              "add, remove and edit mods",
	"perm_edit_mod_visibility":    "edit mod visibility",
	"perm_edit_bw":                "edit banned content",
	"perm_edit_restrictions":      "edit chat restrictions",
	"perm_edit_group":             "edit group",
	"perm_see_counter":            "see counter",
	"perm_see_mod_channel":        "see mod channel",
	"perm_see_mod_actions":        "see mod actions log",
	"perm_can_broadcast":          "send messages in broadcast mode",
	"perm_edit_nlp":               "edit auto-moderation",
	"perm_edit_gp_annc":           "edit group announcement",
	"perm_no_sending_limitations": "bypass message sending limitations",
	"perm_see_ips":                "see IPs",
	"perm_is_staff":               "display staff badge",
	"perm_close_group":            "close group input",
	"perm_unban_all":              "unban all",
	"flood_cont":                  "flood controlled",
	"slow_mode":                   "slow mode restricted to *time* *secs*",
	"second":                      "second",
	"seconds":                     "seconds",
	"disallowed":                  "disallowed",
	"allowed":                     "allowed",
	"enabled":                     "enabled",
	"disabled":                    "disabled",
	"nlp_single_msg":              "nonsense messages (basic)",
	"nlp_msg_queue":               "repetitious messages",
	"nlp_ngram":                   "nonsense messages (advanced)",
	"allow":                       "allow",
	"block":                       "block",
}

var (
	NameFontTag = regexp.MustCompile(`<[nf]\s?[^>]*>`)
)

// ModAction represents a moderation action.
type ModAction struct {
	ID     int       // ID is the unique identifier of the moderation action.
	Type   string    // Type is the type of the moderation action, e.g., "emod", "brdc", "cinm", "chan", "cntr", etc.
	User   string    // User is the name of the moderator who performed the action.
	IP     string    // IP is the IP address of the moderator who performed the action.
	Target string    // Target is the name of the user that the action was performed on.
	Time   time.Time // Time is the timestamp when the action was performed.
	Extra  string    // Extra is any additional information or context related to the moderation action.
}

// ExtraAsSliceInt returns the Extra field as a slice of int64.
//
// Returns:
//   - []int64: The parsed slice of int64 values.
func (ma *ModAction) ExtraAsSliceInt() (ret []int64) {
	json.Unmarshal([]byte(ma.Extra), &ret)
	return
}

// ExtraAsInt returns the Extra field as an int64.
//
// Returns:
//   - int64: The parsed int64 value.
func (ma *ModAction) ExtraAsInt() (ret int64) {
	json.Unmarshal([]byte(ma.Extra), &ret)
	return
}

// ExtraAsBool returns the Extra field as a boolean.
//
// Returns:
//   - bool: The parsed boolean value.
func (ma *ModAction) ExtraAsBool() (ret bool) {
	json.Unmarshal([]byte(ma.Extra), &ret)
	return
}

// ExtraAsSliceInterface returns the Extra field as a slice of interface{}.
//
// Returns:
//   - []interface{}: The parsed slice of interface{} values.
func (ma *ModAction) ExtraAsSliceInterface() (ret []interface{}) {
	json.Unmarshal([]byte(ma.Extra), &ret)
	return
}

// ExtraBanWord returns the Extra field as a [BanWord].
//
// Returns:
//   - BanWord: The parsed BanWord object.
func (ma *ModAction) ExtraBanWord() (ret BanWord) {
	json.Unmarshal([]byte(ma.Extra), &ret)
	return
}

// ExtraDescription returns the Extra field as a [GroupInfo].
//
// Returns:
//   - GroupInfo: The parsed GroupInfo object.
func (ma *ModAction) ExtraDescription() (ret GroupInfo) {
	json.Unmarshal([]byte(ma.Extra), &ret)
	return
}

// String returns a string representation of the [ModAction].
//
// TODO:
//   - Use [strings.Builder] instead.
//
// Returns:
//   - string: The string representation of the ModAction.
func (ma *ModAction) String() (actionDesc string) {
	actionDesc = ModactionTmpl["action_desc_"+ma.Type]
	switch ma.Type {
	case "emod":
		permissions := ma.ExtraAsSliceInt()
		addedPermissions := []string{}
		removedPermissions := []string{}
		var oldFlag, newFlag int64

		// Create a slice of key-value pairs
		pairs := make([]struct {
			key string
			val int64
		}, 0, len(GroupPermissions))

		for key, value := range GroupPermissions {
			pairs = append(pairs, struct {
				key string
				val int64
			}{
				key: key,
				val: value,
			})
		}

		// Sort the slice based on the values
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].val < pairs[j].val
		})

		// Iterate over the sorted slice
		for _, pair := range pairs {
			if pair.val == 131072 || pair.val == 524288 {
				continue
			}
			oldFlag = pair.val & permissions[0]
			newFlag = pair.val & permissions[1]
			if newFlag != oldFlag {
				pair.key = ModactionTmpl["perm_"+strings.ToLower(pair.key)]
				if newFlag != 0 {
					addedPermissions = append(addedPermissions, pair.key)
				} else {
					removedPermissions = append(removedPermissions, pair.key)
				}
			}
		}

		if len(addedPermissions) > 0 {
			actionDesc += " " + strings.Replace(ModactionTmpl["added_perm"], "*added*", strings.Join(addedPermissions, ", "), 1)
		}

		if len(addedPermissions) > 0 && len(removedPermissions) > 0 {
			actionDesc += " and"
		}

		if len(removedPermissions) > 0 {
			actionDesc += " " + strings.Replace(ModactionTmpl["removed_perm"], "*removed*", strings.Join(removedPermissions, ", "), 1)
		}

	case "anon", "prxy":
		allowed := ma.ExtraAsBool()
		if allowed {
			actionDesc = strings.Replace(actionDesc, "*didallow*", ModactionTmpl["allowed"], 1)
		} else {
			actionDesc = strings.Replace(actionDesc, "*didallow*", ModactionTmpl["disallowed"], 1)
		}

	case "brdc", "cinm", "chan", "cntr":
		enabled := ma.ExtraAsBool()
		if enabled {
			actionDesc = strings.Replace(actionDesc, "*didenable*", ModactionTmpl["enabled"], 1)
		} else {
			actionDesc = strings.Replace(actionDesc, "*didenable*", ModactionTmpl["disabled"], 1)
		}

	case "chrl":
		extra := ma.ExtraAsInt()
		if extra == 0 {
			actionDesc = strings.Replace(actionDesc, "*rl*", ModactionTmpl["flood_cont"], 1)
		} else {
			actionDesc = strings.Replace(actionDesc, "*rl*", ModactionTmpl["slow_mode"], 1)
			actionDesc = strings.Replace(actionDesc, "*time*", fmt.Sprintf("%d", extra), 1)
			if extra == 1 {
				actionDesc = strings.Replace(actionDesc, "*secs*", ModactionTmpl["second"], 1)
			} else {
				actionDesc = strings.Replace(actionDesc, "*secs*", ModactionTmpl["seconds"], 1)
			}
		}

	case "annc":
		extra := ma.ExtraAsSliceInterface()
		if extra[0].(float64) > 0 {
			actionDesc = strings.Replace(actionDesc, "*didenable*", ModactionTmpl["enabled"], 1)
			announcement, _ := url.QueryUnescape(extra[2].(string))
			announcement = NameFontTag.ReplaceAllString(announcement, "")
			announcement = strings.ReplaceAll(announcement, "&nbsp;", " ")
			actionDesc += " " + ModactionTmpl["enable_annc"]
			actionDesc = strings.Replace(actionDesc, "*n*", fmt.Sprintf("%d", int(extra[1].(float64))), 1)
			actionDesc = strings.Replace(actionDesc, "*msg*", announcement, 1)
		} else {
			actionDesc = strings.Replace(actionDesc, "*didenable*", ModactionTmpl["disabled"], 1)
		}

	case "enlp":
		permissions := ma.ExtraAsSliceInt()
		allowed := []string{}
		blocked := []string{}
		var oldFlag, newFlag int64
		for _, perms := range []struct {
			name string
			flag int64
		}{
			{"nlp_msg_queue", 32768},
			{"nlp_single_msg", 16384},
			{"nlp_ngram", 2097152},
		} {
			oldFlag = perms.flag & permissions[0]
			newFlag = perms.flag & permissions[1]
			if newFlag != oldFlag {
				if newFlag != 0 {
					allowed = append(allowed, ModactionTmpl[perms.name])
				} else {
					blocked = append(blocked, ModactionTmpl[perms.name])
				}
			}
		}

		if len(allowed) > 0 {
			actionDesc += " " + ModactionTmpl["allow"] + " " + strings.Join(allowed, " and ") + " "
		}

		if len(allowed) > 0 && len(blocked) > 0 {
			actionDesc += " " + ModactionTmpl["and"]
		}

		if len(blocked) > 0 {
			actionDesc += " " + ModactionTmpl["block"] + " " + strings.Join(blocked, " and ") + "."
		}
	}

	actionDesc = strings.Replace(actionDesc, "*name*", ma.User, 1)
	actionDesc = strings.Replace(actionDesc, "*ip*", " ("+ma.IP+")", 1)

	if ma.Target != "" {
		actionDesc = strings.ReplaceAll(actionDesc, "*target*", ma.Target)
	}

	return
}

// ParseModActions parses a string data and returns a slice of [ModAction] objects
//
// Args:
//   - data: The string containing semicolon-separated entries of mod actions.
//
// Returns:
//   - []*ModAction: A slice of pointers to the parsed ModAction objects.
func ParseModActions(data string) (modActions []*ModAction) {
	var (
		fields []string
		id     int
		t      time.Time
		ma     *ModAction
	)

	entries := strings.Split(data, ";")

	for _, entry := range entries {
		fields = strings.SplitN(entry, ",", 7)
		id, _ = strconv.Atoi(fields[0])
		t, _ = utils.ParseTime(fields[5])
		ma = &ModAction{
			ID:     id,
			Type:   fields[1],
			User:   fields[2],
			IP:     fields[3],
			Target: fields[4],
			Time:   t,
			Extra:  fields[6],
		}

		// Append the ModAction to the slice
		modActions = append(modActions, ma)
	}

	return
}
