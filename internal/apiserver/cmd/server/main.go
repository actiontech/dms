//go:generate swagger generate spec
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/actiontech/dms/internal/apiserver/conf"
	"github.com/actiontech/dms/internal/apiserver/service"
	dmsConf "github.com/actiontech/dms/internal/dms/conf"
	pkgLog "github.com/actiontech/dms/internal/pkg/log"
	"github.com/actiontech/dms/pkg/dms-common/pkg/aes"
	"github.com/actiontech/dms/pkg/dms-common/pkg/http"
	"github.com/actiontech/dms/pkg/rand"

	utilIo "github.com/actiontech/dms/pkg/dms-common/pkg/io"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	kLog "github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	Name     = "dms.apiserver"
	pidFile  = "dms.pid"
	Version  string
	flagconf string
)

func init() {
	flag.StringVar(&flagconf, "conf", "config.yaml", "config path, eg: -conf config.yaml")
	dmsConf.Version = Version
}

func run(logger utilLog.Logger) error {
	log_ := utilLog.NewHelper(logger, utilLog.WithMessageKey(Name))

	gctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, errCtx := errgroup.WithContext(gctx)

	opts, err := conf.ReadOptions(logger, flagconf)
	if nil != err {
		return fmt.Errorf("failed to read options: %v", err)
	}
	err = rand.InitSnowflake(opts.ID)
	if nil != err {
		return fmt.Errorf("failed to Init snowflake: %v", err)
	}

	// reset jwt singing key, default dms token
	if err = http.ResetJWTSigningKeyAndDefaultToken(opts.SecretKey); err != nil {
		return err
	}

	// reset aes secret key
	if err = aes.ResetAesSecretKey(opts.SecretKey); err != nil {
		return err
	}

	server, err := service.NewAPIServer(logger, opts)
	if nil != err {
		return fmt.Errorf("failed to new apiserver: %v", err)
	}

	// start server
	g.Go(func() error {
		if err := server.RunHttpServer(logger); nil != err {
			return fmt.Errorf("failed to run http server: %v", err)
		}
		return nil
	})

	g.Go(func() error {
		service.StartAllCronJob(server, errCtx)
		return nil
	})

	// handle exit signal
	g.Go(func() error {

		exit := make(chan os.Signal, 1)
		signal.Notify(exit, syscall.SIGTERM, syscall.SIGINT)
		select {
		case <-errCtx.Done():
			return fmt.Errorf("unexpected shutdown because: %v", errCtx.Err())
		case <-exit:
			log_.Info("shutdown because of signal")

			service.StopAllCronJob()

			if err := server.Shutdown(); nil != err {
				return fmt.Errorf("failed to shutdown: %v", err)
			}

			return nil
		}
	})

	err = startPid(logger, pidFile)
	if nil != err {
		return fmt.Errorf("startPid err: %v", err)
	}
	defer func() {
		if err := stopPid(logger, pidFile); err != nil {
			log_.Errorf("stopPid error: %v", err.Error())
		}
	}()

	if err := g.Wait(); nil != err {
		return err
	}
	log_.Info("shutdown success")
	return nil
}

func main() {
	flag.Parse()

	logger := kLog.With(pkgLog.NewStdLogger(os.Stdout, pkgLog.LogTimeLayout),
		"caller", kLog.DefaultCaller,
	)

	if err := run(pkgLog.NewUtilLogWrapper(logger)); nil != err {
		kLog.NewHelper(kLog.With(logger, "module", Name)).Fatalf("failed to run: %v", err)
	}

}

func startPid(log utilLog.Logger, pidFile string) error {
	return utilIo.WriteFile(log, pidFile, fmt.Sprintf("%v", os.Getpid()), "", 0640)
}

func stopPid(log utilLog.Logger, pidFile string) error {
	return utilIo.Remove(log, pidFile)
}
