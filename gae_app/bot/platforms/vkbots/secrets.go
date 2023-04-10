package vkbots

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/platforms/vk"
)

var BotsBy = botsfw.NewBotSettingsBy(nil,
	vk.NewVkBot(
		strongo.EnvLocal,
		"iframe",
		"5744136", // Local iframe
		"RtoBEsz8gTpWkXrYYe8e",
		"",
		trans.SupportedLocalesByCode5[strongo.LocalCodeRuRu],
	),
	vk.NewVkBot(
		strongo.EnvLocal,
		"site",
		"5709327",
		"t8pWXsRQcL1HSENAGtA6",
		"",
		trans.SupportedLocalesByCode5[strongo.LocalCodeRuRu],
	),
	vk.NewVkBot(
		strongo.EnvDevTest,
		"iframe",
		"5764961",
		"nzAkJmOZqHe5BXHorJ35",
		"",
		trans.SupportedLocalesByCode5[strongo.LocalCodeRuRu],
	),
	vk.NewVkBot(
		strongo.EnvProduction,
		"iframe",
		"5764562",
		"MMtMQJvwfRhvSc0SoLbC",
		common.GA_TRACKING_ID,
		trans.SupportedLocalesByCode5[strongo.LocalCodeRuRu],
	),
)
