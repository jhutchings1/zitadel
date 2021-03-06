package middleware

import (
	"context"

	"github.com/caos/logging"
	"github.com/caos/zitadel/internal/i18n"
	"github.com/rakyll/statik/fs"
	"golang.org/x/text/language"
)

type localizers interface {
	Localizers() []Localizer
}
type Localizer interface {
	LocalizationKey() string
	SetLocalizedMessage(string)
}

func translateFields(ctx context.Context, object localizers, translator *i18n.Translator) {
	if translator == nil || object == nil {
		return
	}
	for _, field := range object.Localizers() {
		field.SetLocalizedMessage(translator.LocalizeFromCtx(ctx, field.LocalizationKey(), nil))
	}
}

func newZitadelTranslator(defaultLanguage language.Tag) *i18n.Translator {
	return translatorFromNamespace("zitadel", defaultLanguage)
}

func translatorFromNamespace(namespace string, defaultLanguage language.Tag) *i18n.Translator {
	dir, err := fs.NewWithNamespace(namespace)
	logging.LogWithFields("ERROR-7usEW", "namespace", namespace).OnError(err).Panic("unable to get namespace")

	translator, err := i18n.NewTranslator(dir, i18n.TranslatorConfig{DefaultLanguage: defaultLanguage})
	logging.Log("ERROR-Sk8sf").OnError(err).Panic("unable to get i18n translator")

	return translator
}
