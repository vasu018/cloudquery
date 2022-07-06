package cmd

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/cloudquery/cloudquery/internal/analytics"
	"github.com/cloudquery/cloudquery/pkg/config"
	"github.com/getsentry/sentry-go"
)

func setAnalyticsProperties(props map[string]interface{}) {
	sprops := make(map[string]string, len(props))
	for k, v := range props {
		analytics.SetGlobalProperty(k, v)
		sprops[k] = fmt.Sprintf("%v", v)
	}
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTags(sprops)
	})
}

func setUserId(newId string) {
	analytics.SetUserId(newId)
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID: newId,
		})
	})
}

func setConfigAnalytics(cfg *config.Config) {
	cfgJSON, _ := json.Marshal(cfg)
	s := sha256.New()
	_, _ = s.Write(cfgJSON)
	cfgHash := fmt.Sprintf("%0x", s.Sum(nil))
	analytics.SetGlobalProperty("cfghash", cfgHash)

	const cfgf = "yaml"
	analytics.SetGlobalProperty("cfgformat", cfgf)

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		if analytics.IsCI() {
			scope.SetUser(sentry.User{
				ID: cfgHash,
			})
		}
		scope.SetTags(map[string]string{
			"cfghash":   cfgHash,
			"cfgformat": cfgf,
		})
	})
}
