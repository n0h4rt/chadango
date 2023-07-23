package chadango

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

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
func (ma *ModAction) ExtraAsSliceInt() (ret []int64) {
	json.Unmarshal([]byte(ma.Extra), &ret)
	return
}

// ExtraAsInt returns the Extra field as an int64.
func (ma *ModAction) ExtraAsInt() (ret int64) {
	json.Unmarshal([]byte(ma.Extra), &ret)
	return
}

// ExtraAsBool returns the Extra field as a boolean.
func (ma *ModAction) ExtraAsBool() (ret bool) {
	json.Unmarshal([]byte(ma.Extra), &ret)
	return
}

// ExtraAsSliceInterface returns the Extra field as a slice of interface{}.
func (ma *ModAction) ExtraAsSliceInterface() (ret []interface{}) {
	json.Unmarshal([]byte(ma.Extra), &ret)
	return
}

// ExtraBanWord returns the Extra field as a BanWord.
func (ma *ModAction) ExtraBanWord() (ret BanWord) {
	json.Unmarshal([]byte(ma.Extra), &ret)
	return
}

// ExtraDescription returns the Extra field as a GroupInfo.
func (ma *ModAction) ExtraDescription() (ret GroupInfo) {
	json.Unmarshal([]byte(ma.Extra), &ret)
	return
}

// String returns a string representation of the ModAction.
// TODO: use `strings.Builder` instead.
func (ma *ModAction) String() (actionDesc string) {
	actionDesc = ModactionTmpl["action_desc_"+ma.Type]
	switch ma.Type {
	case "emod":
		permissions := ma.ExtraAsSliceInt()
		addedPermissions := []string{}
		removedPermissions := []string{}
		var oldFlag, newFlag int64

		// Create a slice of key-value pairs
		pairs := make([]KeyValue, 0, len(GroupPermissions))
		for key, value := range GroupPermissions {
			pairs = append(pairs, KeyValue{Key: key, Value: value})
		}

		// Sort the slice based on the values
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Value < pairs[j].Value
		})

		// Iterate over the sorted slice
		for _, pair := range pairs {
			if pair.Value == 131072 || pair.Value == 524288 {
				continue
			}
			oldFlag = pair.Value & permissions[0]
			newFlag = pair.Value & permissions[1]
			if newFlag != oldFlag {
				pair.Key = ModactionTmpl["perm_"+strings.ToLower(pair.Key)]
				if newFlag != 0 {
					addedPermissions = append(addedPermissions, pair.Key)
				} else {
					removedPermissions = append(removedPermissions, pair.Key)
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
		var permissionName string
		var flag, oldFlag, newFlag int64
		for permissionName, flag = range map[string]int64{
			"nlp_msg_queue":  32768,
			"nlp_single_msg": 16384,
			"nlp_ngram":      2097152,
		} {
			oldFlag = flag & permissions[0]
			newFlag = flag & permissions[1]
			if newFlag != oldFlag {
				permissionName = ModactionTmpl[permissionName]
				if newFlag != 0 {
					allowed = append(allowed, permissionName)
				} else {
					blocked = append(blocked, permissionName)
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

// ParseModActions parses a string data and returns a slice of ModAction objects
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
		t, _ = ParseTime(fields[5])
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
