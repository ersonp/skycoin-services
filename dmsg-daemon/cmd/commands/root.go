package commands

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	proxyproto "github.com/pires/go-proxyproto"

	"github.com/skycoin/dmsg/buildinfo"
	"github.com/skycoin/dmsg/cmdutil"
	"github.com/skycoin/skycoin-services/dmsg-daemon/cmd/internal/api"

	"github.com/spf13/cobra"
)

const defaultTick = 60 * time.Second

var (
	sf   cmdutil.ServiceFlags
	addr string
	tick time.Duration
)

func init() {
	sf.Init(rootCmd, "dmsg_daemon", "")

	rootCmd.Flags().StringVarP(&addr, "addr", "a", ":9090", "address to bind to")
	rootCmd.Flags().DurationVar(&tick, "entry-timeout", defaultTick, "discovery entry timeout")
}

var rootCmd = &cobra.Command{
	Use:   "dmsg-daemon",
	Short: "Dmsg daemon service",
	Run: func(_ *cobra.Command, _ []string) {
		if _, err := buildinfo.Get().WriteTo(os.Stdout); err != nil {
			log.Printf("Failed to output build info: %v", err)
		}

		log := sf.Logger()

		a := api.New(log)

		ctx, cancel := cmdutil.SignalContext(context.Background(), log)
		defer cancel()
		log.WithField("addr", addr).Info("Serving discovery API...")
		go func() {
			if err := listenAndServe(addr, a); err != nil {
				log.Errorf("ListenAndServe: %v", err)
				cancel()
			}
		}()
		<-ctx.Done()

	},
}

// Execute executes root CLI command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func listenAndServe(addr string, handler http.Handler) error {
	srv := &http.Server{Addr: addr, Handler: handler}
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	proxyListener := &proxyproto.Listener{Listener: ln}
	defer proxyListener.Close() // nolint:errcheck
	if err != nil {
		return err
	}
	return srv.Serve(proxyListener)
}
