package vkbots

import (
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/vk"
)

var BotsBy = bots.NewBotSettingsBy(nil,
	vk.NewVkBot(
		strongo.EnvLocal,
		"iframe",
		"5744136", // Local iframe
		"RtoBEsz8gTpWkXrYYe8e",
		trans.SupportedLocalesByCode5[strongo.LocalCodeRuRu],
	),
	vk.NewVkBot(
		strongo.EnvLocal,
		"site",
		"5709327",
		"t8pWXsRQcL1HSENAGtA6",
		trans.SupportedLocalesByCode5[strongo.LocalCodeRuRu],
	),
	vk.NewVkBot(
		strongo.EnvDevTest,
		"iframe",
		"5764961",
		"nzAkJmOZqHe5BXHorJ35",
		trans.SupportedLocalesByCode5[strongo.LocalCodeRuRu],
	),
	vk.NewVkBot(
		strongo.EnvProduction,
		"iframe",
		"5764562",
		"MMtMQJvwfRhvSc0SoLbC",
		trans.SupportedLocalesByCode5[strongo.LocalCodeRuRu],
	),
)
