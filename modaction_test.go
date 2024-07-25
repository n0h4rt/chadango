package chadango

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestModAction_ExtraAsSliceInt(t *testing.T) {
	ma := ParseModActions("6397575,enlp,perorist,127.0.0.1,None,1688488704,[2113536,0]")[0]
	expected := []int64{2113536, 0}
	result := ma.ExtraAsSliceInt()
	assert.Equal(t, expected, result, "ExtraAsSliceInt() should return the expected slice")
}

func TestModAction_ExtraAsInt(t *testing.T) {
	ma := ParseModActions("6397571,chrl,perorist,127.0.0.1,None,1688488618,30")[0]
	expected := int64(30)
	result := ma.ExtraAsInt()
	assert.Equal(t, expected, result, "ExtraAsInt() should return the expected int")
}

func TestModAction_ExtraAsBool(t *testing.T) {
	ma := ParseModActions("6397569,prxy,perorist,127.0.0.1,None,1688488541,true")[0]
	expected := true
	result := ma.ExtraAsBool()
	assert.Equal(t, expected, result, "ExtraAsBool() should return the expected bool")
}

func TestModAction_ExtraAsSliceInterface(t *testing.T) {
	ma := ParseModActions(`6397553,annc,perorist,127.0.0.1,None,1688488200,[0,60,"%3Cnf60/%3E%3Cf%20x000%3D%22%22%3E3%26nbsp%3Bkata%26nbsp%3Bterserah"]`)[0]
	expected := []interface{}{0.0, 60.0, "%3Cnf60/%3E%3Cf%20x000%3D%22%22%3E3%26nbsp%3Bkata%26nbsp%3Bterserah"}
	result := ma.ExtraAsSliceInterface()
	assert.Equal(t, expected, result, "ExtraAsSliceInterface() should return the expected slice")
}

func TestModAction_ExtraBanWord(t *testing.T) {
	ma := &ModAction{
		Extra: `{"words":"example","wholeWords":"text"}`,
	}
	expected := BanWord{Words: "example", WholeWords: "text"}
	result := ma.ExtraBanWord()
	assert.Equal(t, expected, result, "ExtraBanWord() should return the expected BanWord")
}

func TestModAction_ExtraDescription(t *testing.T) {
	ma := &ModAction{
		Extra: `{"title":"Group Name","ownr_msg":"Group Description"}`,
	}
	expected := GroupInfo{Title: "Group Name", OwnerMessage: "Group Description"}
	result := ma.ExtraDescription()
	assert.Equal(t, expected, result, "ExtraDescription() should return the expected GroupInfo")
}

func TestModAction_String(t *testing.T) {
	ma0 := ParseModActions(
		`6399304,emod,perorist,127.0.0.1,perorist,1688700775,[353224,1402876];` +
			`6401361,anon,perorist,127.0.0.1,None,1688925129,true;` +
			`6397575,enlp,perorist,127.0.0.1,None,1688488704,[2113536,0];` +
			`6401364,annc,perorist,127.0.0.1,None,1688925246,[1,60,"%3CnA149A0/%3E%3Cf%20xf9f%3D%22%22%3E3%26nbsp%3Bkata%26nbsp%3Bterserah"];` +
			`6397598,egrp,perorist,127.0.0.1,None,1688489707,{"ownr_msg":"Heathens%20come%20join","title":"Only%20me"}`,
	)
	ma1 := ma0[0]
	expected1 := "perorist (127.0.0.1) gave perorist permission to: edit mod visibility, edit chat restrictions, edit group, edit group announcement, unban all"
	result1 := ma1.String()
	assert.Equal(t, expected1, result1, "String() should return the expected string for emod type")

	ma2 := ma0[1]
	expected2 := "perorist (127.0.0.1) allowed anons in the group"
	result2 := ma2.String()
	assert.Equal(t, expected2, result2, "String() should return the expected string for anon type")

	ma3 := ma0[2]
	expected3 := "perorist (127.0.0.1) changed auto-moderation to block nonsense messages (basic) and nonsense messages (advanced)."
	result3 := ma3.String()
	assert.Equal(t, expected3, result3, "String() should return the expected string for chrl type")

	ma4 := ma0[3]
	expected4 := "perorist (127.0.0.1) enabled auto-announcement every 60 seconds: 3 kata terserah"
	result4 := ma4.String()
	assert.Equal(t, expected4, result4, "String() should return the expected string for annc type")

	ma5 := ma0[4]
	expected5 := "perorist (127.0.0.1) edited group"
	result5 := ma5.String()
	assert.Equal(t, expected5, result5, "String() should return the expected string for enlp type")
}

func TestParseModActions(t *testing.T) {
	data := "1,emod,user1,127.0.0.1,target1,1688488704,[1,2,3];2,anon,user2,127.0.0.1,target2,1688488704,true"
	expected := []*ModAction{
		{
			ID:     1,
			Type:   "emod",
			User:   "user1",
			IP:     "127.0.0.1",
			Target: "target1",
			Time:   time.Date(2023, time.July, 4, 16, 38, 24, 0, time.UTC).Local(),
			Extra:  "[1,2,3]",
		},
		{
			ID:     2,
			Type:   "anon",
			User:   "user2",
			IP:     "127.0.0.1",
			Target: "target2",
			Time:   time.Date(2023, time.July, 4, 16, 38, 24, 0, time.UTC).Local(),
			Extra:  "true",
		},
	}
	result := ParseModActions(data)
	assert.Equal(t, *expected[0], *result[0], "ParseModActions() should return the expected ModAction slice")
	assert.Equal(t, *expected[1], *result[1], "ParseModActions() should return the expected ModAction slice")
}
