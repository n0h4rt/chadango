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
	assert.Equal(t, expected3, result3, "String() should return the expected string for enlp type")

	ma4 := ma0[3]
	expected4 := "perorist (127.0.0.1) enabled auto-announcement every 60 seconds: 3 kata terserah"
	result4 := ma4.String()
	assert.Equal(t, expected4, result4, "String() should return the expected string for annc type")

	ma5 := ma0[4]
	expected5 := "perorist (127.0.0.1) edited group"
	result5 := ma5.String()
	assert.Equal(t, expected5, result5, "String() should return the expected string for egrp type")
}

func TestParseModActions_UniqueTypes(t *testing.T) {
	actions := ParseModActions(
		`6415678,chbw,perorist,127.0.0.1,None,1690621738,{"wholeWords":"","words":""};` +
			`6413806,emod,clonerxyz,127.0.0.1,metia,1690388088,[1173496,1139656];` +
			`6405918,prxy,perorist,127.0.0.1,None,1689441854,true;` +
			`6401364,annc,perorist,127.0.0.1,None,1688925246,[0,60,"%3CnA149A0/%3E%3Cf%20xf9f%3D%22%22%3E3%26nbsp%3Bkata%26nbsp%3Bterserah"];` +
			`6401361,anon,perorist,127.0.0.1,None,1688925129,true;` +
			`6397604,cntr,clonerxyz,127.0.0.1,None,1688490165,true;` +
			`6397598,egrp,clonerxyz,127.0.0.1,None,1688489707,{"ownr_msg":"Heathens%20come%20join","title":"Only%20me"};` +
			`6397594,cinm,clonerxyz,127.0.0.1,None,1688489415,false;` +
			`6397590,aopn,None,None,None,1688489386,;` +
			`6397588,acls,None,None,None,1688489305,;` +
			`6397582,chsi,clonerxyz,127.0.0.1,None,1688488891,;` +
			`6397581,hidi,clonerxyz,127.0.0.1,None,1688488881,;` +
			`6397580,shwi,clonerxyz,127.0.0.1,None,1688488829,;` +
			`6397579,enlp,clonerxyz,127.0.0.1,None,1688488814,[0,16384];` +
			`6397574,chrl,clonerxyz,127.0.0.1,None,1688488651,0`,
	)

	expected := []*ModAction{
		{ID: 6415678, Type: "chbw", User: "perorist", IP: "127.0.0.1", Target: "None", Time: time.Unix(1690621738, 0), Extra: `{"wholeWords":"","words":""}`},
		{ID: 6413806, Type: "emod", User: "clonerxyz", IP: "127.0.0.1", Target: "metia", Time: time.Unix(1690388088, 0), Extra: `[1173496,1139656]`},
		{ID: 6405918, Type: "prxy", User: "perorist", IP: "127.0.0.1", Target: "None", Time: time.Unix(1689441854, 0), Extra: "true"},
		{ID: 6401364, Type: "annc", User: "perorist", IP: "127.0.0.1", Target: "None", Time: time.Unix(1688925246, 0), Extra: `[0,60,"%3CnA149A0/%3E%3Cf%20xf9f%3D%22%22%3E3%26nbsp%3Bkata%26nbsp%3Bterserah"]`},
		{ID: 6401361, Type: "anon", User: "perorist", IP: "127.0.0.1", Target: "None", Time: time.Unix(1688925129, 0), Extra: "true"},
		{ID: 6397604, Type: "cntr", User: "clonerxyz", IP: "127.0.0.1", Target: "None", Time: time.Unix(1688490165, 0), Extra: "true"},
		{ID: 6397598, Type: "egrp", User: "clonerxyz", IP: "127.0.0.1", Target: "None", Time: time.Unix(1688489707, 0), Extra: `{"ownr_msg":"Heathens%20come%20join","title":"Only%20me"}`},
		{ID: 6397594, Type: "cinm", User: "clonerxyz", IP: "127.0.0.1", Target: "None", Time: time.Unix(1688489415, 0), Extra: "false"},
		{ID: 6397590, Type: "aopn", User: "None", IP: "None", Target: "None", Time: time.Unix(1688489386, 0), Extra: ""},
		{ID: 6397588, Type: "acls", User: "None", IP: "None", Target: "None", Time: time.Unix(1688489305, 0), Extra: ""},
		{ID: 6397582, Type: "chsi", User: "clonerxyz", IP: "127.0.0.1", Target: "None", Time: time.Unix(1688488891, 0), Extra: ""},
		{ID: 6397581, Type: "hidi", User: "clonerxyz", IP: "127.0.0.1", Target: "None", Time: time.Unix(1688488881, 0), Extra: ""},
		{ID: 6397580, Type: "shwi", User: "clonerxyz", IP: "127.0.0.1", Target: "None", Time: time.Unix(1688488829, 0), Extra: ""},
		{ID: 6397579, Type: "enlp", User: "clonerxyz", IP: "127.0.0.1", Target: "None", Time: time.Unix(1688488814, 0), Extra: `[0,16384]`},
		{ID: 6397574, Type: "chrl", User: "clonerxyz", IP: "127.0.0.1", Target: "None", Time: time.Unix(1688488651, 0), Extra: "0"},
	}

	if len(actions) != len(expected) {
		t.Fatalf("Expected %d actions, but got %d", len(expected), len(actions))
	}

	for i, action := range actions {
		if *action != *expected[i] {
			t.Errorf("ModAction %d mismatch. Expected %+v, but got %+v", i, *expected[i], *action)
		}

		if action.String() != expected[i].String() {
			t.Errorf("String() %d mismatch. Expected %+v, but got %+v", i, expected[i].String(), action.String())
		}
	}
}
