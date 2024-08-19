package locale

import (
	"embed"
	"fmt"
	"github.com/BurntSushi/toml"
	//"github.com/actiontech/sqle/sqle/log"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed active.*.toml
var LocaleFS embed.FS

var bundle *i18n.Bundle

var newEntry log

type log interface {
	Error(...interface{})
}

func MustInit(l log) {
	newEntry = l
	bundle = i18n.NewBundle(language.Chinese)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	_, err := bundle.LoadMessageFileFS(LocaleFS, "active.zh.toml")
	if err != nil {
		panic(fmt.Sprintf("load i18n config failed, error: %v", err))
	}
	_, err = bundle.LoadMessageFileFS(LocaleFS, "active.en.toml")
	if err != nil {
		panic(fmt.Sprintf("load i18n config failed, error: %v", err))
	}
}

func I18nEchoMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accept := c.Request().Header.Get("Accept-Language")
			localizer := i18n.NewLocalizer(bundle, accept)

			c.Set("localizer", localizer)
			return next(c)
		}
	}
}

func MustGetLocaleFromCtx(c echo.Context) *i18n.Localizer {
	localizer, ok := c.Get("localizer").(*i18n.Localizer)
	if !ok {
		return i18n.NewLocalizer(bundle)
	}
	return localizer
}

func GetLocalizerByAcceptLanguage(c echo.Context) *i18n.Localizer {
	accept := c.Request().Header.Get("Accept-Language")
	return i18n.NewLocalizer(bundle, accept)
}

func ShouldLocalizeMsg(localizer *i18n.Localizer, msg *i18n.Message) string {
	m, err := localizer.LocalizeMessage(msg)
	if err != nil && newEntry != nil {
		newEntry.Error("LocalizeMessage:", msg, "failed:", err)
	}
	return m
}
