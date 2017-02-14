package translations

import (
	"strings"
	"github.com/nicksnyder/go-i18n/i18n"
)

var B i18n.TranslateFunc
var Lang = "en-US"

func T(text string) string{
        if B == nil {
                i18n.LoadTranslationFile("translations/" + strings.ToLower(Lang) + ".all.json")
                B, _ = i18n.Tfunc(Lang)
        }

        return B(text)
}

func ChangeLang(lang string) {
        Lang = lang
        B = nil
}

