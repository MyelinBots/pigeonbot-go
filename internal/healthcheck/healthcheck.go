package healthcheck

import (
	"context"
	"github.com/MyelinBots/pigeonbot-go/config"
	"net/http"
	"strconv"
)

// Healthcheck that starts http server
func StartHealthcheck(ctx context.Context, cfg config.AppConfig) {
	// start http server
	go func() {
		port := strconv.Itoa(cfg.Port)
		err := http.ListenAndServe(":"+port, HealthCheckHandler())
		if err != nil {
			panic(err)
		}
	}()

}

func HealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
