package runner

import (
	"context"

	internalConfig "github.com/wallarm/gotestwaf/internal/config"

	"github.com/wallarm/gotestwaf/pkg/config"
	"github.com/wallarm/gotestwaf/pkg/statistic"

	"github.com/getkin/kin-openapi/routers"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/wallarm/gotestwaf/internal/db"
	"github.com/wallarm/gotestwaf/internal/openapi"
	"github.com/wallarm/gotestwaf/internal/scanner"
	"github.com/wallarm/gotestwaf/internal/version"
)

type GoTestWAFRunner struct {
	cfg *internalConfig.Config
}

func NewGoTestWAFRunner(cfg config.Config) (*GoTestWAFRunner, error) {
	internalCfg := internalConfig.Config(cfg)
	return &GoTestWAFRunner{cfg: &internalCfg}, nil
}

func (g *GoTestWAFRunner) Run(ctx context.Context) (*statistic.Statistics, error) {
	logger := logrus.New()
	logger.WithField("version", version.Version).Info("GoTestWAF started")

	var err error
	var router routers.Router
	var templates openapi.Templates

	logger.Info("Test cases loading started")

	testCases, err := db.LoadTestCases(g.cfg)
	if err != nil {
		return nil, errors.Wrap(err, "loading test case")
	}

	logger.Info("Test cases loading finished")

	db, err := db.NewDB(testCases)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't create test cases DB")
	}

	logger.WithField("fp", db.Hash).Info("Test cases fingerprint")

	logger.WithField("http_client", g.cfg.HTTPClient).
		Infof("%s is used as an HTTP client to make requests", g.cfg.HTTPClient)

	s, err := scanner.New(logger, g.cfg, db, templates, router, g.cfg.AddDebugHeader)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't create scanner")
	}

	if g.cfg.HTTPClient != "chrome" {
		isJsReuqired, err := s.CheckIfJavaScriptRequired(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't check if JavaScript is required to interact with the endpoint")
		}

		if isJsReuqired {
			return nil, errors.New("JavaScript is required to interact with the endpoint")
		}
	}

	if !g.cfg.SkipWAFBlockCheck {
		err = s.WAFBlockCheck(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		logger.WithField("status", "skipped").Info("WAF pre-check")
	}

	s.CheckGRPCAvailability(ctx)
	s.CheckGraphQLAvailability(ctx)

	err = s.Run(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error occurred while scanning")
	}

	stat := db.GetStatistics(g.cfg.IgnoreUnresolved, g.cfg.NonBlockedAsPassed)
	publicStat := stat.ToPublicStatistic()

	return &publicStat, nil
}

/*
func main() {
	logger := logrus.New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-shutdown
		logger.WithField("signal", sig).Info("scan canceled")
		cancel()
	}()

	args, err := parseFlags()
	if err != nil {
		logger.WithError(err).Error("couldn't parse flags")
		os.Exit(1)
	}

	logger.SetLevel(logLevel)
	if logFormat == jsonLogFormat {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}
	if quiet {
		logger.SetOutput(io.Discard)
	}

	cfg, err := loadConfig()
	if err != nil {
		logger.WithError(err).Error("couldn't load config")
		os.Exit(1)
	}

	if !cfg.HideArgsInReport {
		cfg.Args = args
	}

	if err := run(ctx, cfg, logger); err != nil {
		logger.WithError(err).Error("caught error in main function")
		os.Exit(1)
	}
}
*/
