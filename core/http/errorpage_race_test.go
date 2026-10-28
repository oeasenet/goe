package http

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// A kernel built while another one renders its error page must not race on
// the shared template.
func TestNew_KernelsBuiltWhileAnotherServesShareTheErrorPage(t *testing.T) {
	mod, _, addr := newServingModule(t)
	require.NoError(t, mod.OnStart(context.Background()))
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = mod.OnStop(ctx)
	}()

	var wg sync.WaitGroup
	wg.Go(func() {
		for range 20 {
			resp, err := http.Get("http://" + addr() + "/missing") // renders the error page
			if err == nil {
				_ = resp.Body.Close()
			}
		}
	})
	for range 20 {
		_ = New(newFakeConfig(nil), newOptionsLogger())
	}
	wg.Wait()
}
