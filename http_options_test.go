package goe

import (
	"os"
	"testing"

	goehttp "go.oease.dev/goe/v2/core/http"
)

// resetInstance clears the package-level singleton so each subtest starts from
// a clean container, matching the convention in config_override_test.go.
func resetInstance() {
	instance.app = nil
	instance.config = nil
	instance.logger = nil
	instance.http = nil
}

func TestOptions_HTTPCodeFirstConfig(t *testing.T) {
	_ = os.Unsetenv("HTTP_PORT")
	_ = os.Unsetenv("FIBER_SERVER_HEADER")

	t.Run("supplying HTTP options enables the module", func(t *testing.T) {
		defer resetInstance()

		_ = New(Options{
			// Deliberately no WithHTTP: true. Passing options is an
			// unambiguous request for the module.
			HTTP: []goehttp.Option{
				goehttp.WithServerHeader("implied"),
			},
		})

		kernel := HTTP()
		if kernel == nil {
			t.Fatal("expected the HTTP module to be enabled by the presence of HTTP options")
		}
		if got := kernel.App().Config().ServerHeader; got != "implied" {
			t.Errorf("expected options to reach the kernel, got ServerHeader %q", got)
		}
	})

	t.Run("options are threaded through to the fiber app", func(t *testing.T) {
		defer resetInstance()

		_ = New(Options{
			HTTP: []goehttp.Option{
				goehttp.WithServerHeader("from-code"),
				goehttp.WithBodyLimit(9 << 20),
				goehttp.WithStrictRouting(true),
			},
		})

		cfg := HTTP().App().Config()
		if cfg.ServerHeader != "from-code" {
			t.Errorf("expected ServerHeader from-code, got %q", cfg.ServerHeader)
		}
		if cfg.BodyLimit != 9<<20 {
			t.Errorf("expected BodyLimit %d, got %d", 9<<20, cfg.BodyLimit)
		}
		if !cfg.StrictRouting {
			t.Error("expected StrictRouting to be enabled")
		}
	})

	t.Run("code options win over ConfigOverrides", func(t *testing.T) {
		defer resetInstance()

		// ConfigOverrides feed the environment layer, which options sit above.
		_ = New(Options{
			ConfigOverrides: map[string]any{"FIBER_SERVER_HEADER": "from-env"},
			HTTP: []goehttp.Option{
				goehttp.WithServerHeader("from-code"),
			},
		})

		if got := HTTP().App().Config().ServerHeader; got != "from-code" {
			t.Errorf("expected code to win over the environment layer, got %q", got)
		}
	})

	t.Run("environment still fills gaps", func(t *testing.T) {
		defer resetInstance()

		_ = New(Options{
			ConfigOverrides: map[string]any{"FIBER_SERVER_HEADER": "from-env"},
			HTTP: []goehttp.Option{
				goehttp.WithBodyLimit(7 << 20), // unrelated field
			},
		})

		cfg := HTTP().App().Config()
		if cfg.ServerHeader != "from-env" {
			t.Errorf("expected the environment to supply ServerHeader, got %q", cfg.ServerHeader)
		}
		if cfg.BodyLimit != 7<<20 {
			t.Errorf("expected BodyLimit %d, got %d", 7<<20, cfg.BodyLimit)
		}
	})

	t.Run("WithHTTP alone still works", func(t *testing.T) {
		defer resetInstance()

		_ = New(Options{WithHTTP: true})

		if HTTP() == nil {
			t.Fatal("expected the HTTP module to be enabled")
		}
		if got := HTTP().App().Config().ServerHeader; got != "Goe" {
			t.Errorf("expected the GOE default ServerHeader, got %q", got)
		}
	})
}
