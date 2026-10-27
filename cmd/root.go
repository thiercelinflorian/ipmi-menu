package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thiercelinflorian/ipmi-menu/internal/config"
	"github.com/thiercelinflorian/ipmi-menu/internal/ipmi"
	"github.com/thiercelinflorian/ipmi-menu/internal/tui"
)

var (
	flagServer   string
	flagHost     string
	flagUser     string
	flagPassword string
	flagPort     int
)

var rootCmd = &cobra.Command{
	Use:   "ipmi-menu",
	Short: "Piloter des serveurs en IPMI : menu interactif ou commandes directes",
	Long: `Sans argument, ouvre le menu interactif.
Avec une sous-commande, agit directement (scriptable, sans confirmation).

Exemples :
  ipmi-menu                         menu interactif
  ipmi-menu -s srv1                 menu directement sur srv1
  ipmi-menu power on -s srv1
  ipmi-menu boot pxe --reboot -s srv1
  ipmi-menu sol -H 10.0.0.10 -U ADMIN
  ipmi-menu host add srv1 10.0.0.10 -U ADMIN`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ipmi.Check(); err != nil {
			return err
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		var t *ipmi.Target
		if flagServer != "" || flagHost != "" {
			tt, err := resolve(cfg)
			if err != nil {
				return err
			}
			t = &tt
		}
		return tui.Run(cfg, t)
	},
}

func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "erreur :", err)
		os.Exit(1)
	}
}

func init() {
	f := rootCmd.PersistentFlags()
	f.StringVarP(&flagServer, "server", "s", "", "serveur enregistré à utiliser")
	f.StringVarP(&flagHost, "host", "H", "", "adresse du BMC (sans passer par un serveur enregistré)")
	f.StringVarP(&flagUser, "user", "U", "", "utilisateur IPMI (défaut : root)")
	f.StringVarP(&flagPassword, "password", "P", "", "mot de passe IPMI (ou variable IPMI_PASSWORD)")
	f.IntVar(&flagPort, "port", 0, "port RMCP (défaut : 623)")
	rootCmd.RegisterFlagCompletionFunc("server", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		cfg, err := config.Load()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		return cfg.Names(), cobra.ShellCompDirectiveNoFileComp
	})
}

func resolve(cfg *config.Config) (ipmi.Target, error) {
	var t ipmi.Target
	switch {
	case flagHost != "":
		t.Host = flagHost
	case flagServer != "":
		var ok bool
		if t, ok = cfg.Get(flagServer); !ok {
			return t, fmt.Errorf("serveur %q inconnu (serveurs : %s)", flagServer, list(cfg))
		}
	case len(cfg.Hosts) == 1:
		t, _ = cfg.Get(cfg.Names()[0])
	case len(cfg.Hosts) == 0:
		return t, errors.New("aucun serveur : utilise -H ADRESSE ou enregistre-en un avec `ipmi-menu host add`")
	default:
		return t, fmt.Errorf("précise le serveur avec -s (serveurs : %s)", list(cfg))
	}
	if flagUser != "" {
		t.User = flagUser
	}
	if p := passwordFlag(); p != "" {
		t.Password = p
	}
	if flagPort != 0 {
		t.Port = flagPort
	}
	return t, nil
}

func passwordFlag() string {
	if flagPassword != "" {
		return flagPassword
	}
	return os.Getenv("IPMI_PASSWORD")
}

func list(cfg *config.Config) string { return strings.Join(cfg.Names(), ", ") }

func target() (ipmi.Target, error) {
	if err := ipmi.Check(); err != nil {
		return ipmi.Target{}, err
	}
	cfg, err := config.Load()
	if err != nil {
		return ipmi.Target{}, err
	}
	return resolve(cfg)
}

func output(out string, err error) error {
	if out != "" {
		fmt.Println(out)
	}
	return err
}
