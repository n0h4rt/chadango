package models

import (
	"fmt"
	"net/url"
	"reflect"
)

// MessageStyle represents the style settings for a message.
type MessageStyle struct {
	FontFamily    string `json:"fontFamily"`    // The font family used for the message text.
	FontSize      string `json:"fontSize"`      // The font size used for the message text.
	Bold          bool   `json:"bold"`          // A boolean value indicating whether the message text should be displayed in bold.
	StylesOn      bool   `json:"stylesOn"`      // A boolean value indicating whether the message styles are enabled.
	UseBackground string `json:"usebackground"` // The background color used for the message text.
	Italics       bool   `json:"italics"`       // A boolean value indicating whether the message text should be displayed in italics.
	TextColor     string `json:"textColor"`     // The color used for the message text.
	Underline     bool   `json:"underline"`     // A boolean value indicating whether the message text should be underlined.
	NameColor     string `json:"nameColor"`     // The color used for the username or sender's name in the message.
}

// GetForm returns the URL-encoded form values for the `MessageStyle`.
func (mb MessageStyle) GetForm() url.Values {
	form := url.Values{}
	configType := reflect.TypeOf(mb)
	configValue := reflect.ValueOf(mb)
	var (
		field reflect.StructField
		value reflect.Value
		tag   string
	)

	for i := 0; i < configType.NumField(); i++ {
		field = configType.Field(i)
		value = configValue.Field(i)
		tag = field.Tag.Get("json")

		form.Set(tag, fmt.Sprintf("%v", value.Interface()))
	}

	return form
}
