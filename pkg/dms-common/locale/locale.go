package locale

import (
	"context"
	"embed"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

const LocalizerCtxKey = "localizer"

//go:embed active.*.toml
var LocaleFS embed.FS

var bundle *i18n.Bundle

var newEntry log

type log interface {
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
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

func EchoMiddlewareI18nByAcceptLanguage() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			lang := c.Request().Header.Get("Accept-Language")
			localizer := i18n.NewLocalizer(bundle, lang)

			ctx := context.WithValue(c.Request().Context(), LocalizerCtxKey, localizer)
			ctx = context.WithValue(ctx, "Accept-Language", lang)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func ShouldLocalizeMsg(ctx context.Context, msg *i18n.Message) string {
	l, ok := ctx.Value(LocalizerCtxKey).(*i18n.Localizer)
	if !ok {
		l = i18n.NewLocalizer(bundle)
		newEntry.Warnf("No localizer in context when localize msg: %v, use default", msg.ID)
	}

	m, err := l.LocalizeMessage(msg)
	if err != nil {
		newEntry.Errorf("LocalizeMessage: %v failed: %v", msg.ID, err)
	}
	return m
}