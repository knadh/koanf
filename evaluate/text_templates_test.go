package evaluate

import (
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/stretchr/testify/assert"
	"testing"
)

var defaultConfig = map[string]interface{}{
	"app.dir": "/app/path",

	"assets.dir": `{{koanf "app.dir"}}/assets`,

	"notifications.enabled":       true,
	"notifications.templates.dir": `{{koanf "assets.dir"}}/notifications`,
}

func TestTextTemplates(t *testing.T) {
	tests := []struct {
		name    string
		configs []map[string]interface{}
		want    map[string]interface{}
	}{
		{
			"default",
			[]map[string]interface{}{
				defaultConfig,
			},
			map[string]interface{}{
				"app.dir": "/app/path",

				"assets.dir": "/app/path/assets",

				"notifications.enabled":       true,
				"notifications.templates.dir": `/app/path/assets/notifications`,
			},
		},
		{
			"override-app-path",
			[]map[string]interface{}{
				defaultConfig,
				{
					"app.dir": "/opt/app/path",
				},
			},
			map[string]interface{}{
				"app.dir": "/opt/app/path",

				"assets.dir": "/opt/app/path/assets",

				"notifications.enabled":       true,
				"notifications.templates.dir": `/opt/app/path/assets/notifications`,
			},
		},
		{
			"override-assets-path",
			[]map[string]interface{}{
				defaultConfig,
				{
					"assets.dir": "/var/app/assets",
				},
			},
			map[string]interface{}{
				"app.dir": "/app/path",

				"assets.dir": "/var/app/assets",

				"notifications.enabled":       true,
				"notifications.templates.dir": `/var/app/assets/notifications`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := koanf.New(".")
			for _, config := range tt.configs {
				err := k.Load(confmap.Provider(config, ""), nil)
				assert.Nil(t, err)
			}

			evaluetedConfig := TextTemplates(k)
			assert.Equal(t, evaluetedConfig.All(), tt.want)
		})
	}
}
