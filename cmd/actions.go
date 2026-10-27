package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thiercelinflorian/ipmi-menu/internal/ipmi"
)

var (
	bootLegacy     bool
	bootPersistent bool
	bootReboot     bool
)

var powerCmd = &cobra.Command{
	Use:       "power [status|on|off|soft|cycle|reset]",
	Short:     "État ou action d'alimentation (status par défaut)",
	Args:      cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
	ValidArgs: ipmi.PowerActions,
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := target()
		if err != nil {
			return err
		}
		action := "status"
		if len(args) == 1 {
			action = args[0]
		}
		return output(t.Power(action))
	},
}

var bootCmd = &cobra.Command{
	Use:       "boot <pxe|disk|cdrom|bios>",
	Short:     "Choisir le périphérique du prochain démarrage",
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: ipmi.BootDevices,
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := target()
		if err != nil {
			return err
		}
		dropped, err := t.SetBoot(args[0], ipmi.BootOptions{Legacy: bootLegacy, Persistent: bootPersistent})
		if err != nil {
			return err
		}
		fmt.Println("Prochain démarrage :", args[0])
		if dropped {
			fmt.Fprintln(os.Stderr, "attention : options UEFI/persistant refusées par le BMC")
		}
		if bootReboot {
			if err := t.Reboot(); err != nil {
				return err
			}
			fmt.Println("Serveur redémarré")
		}
		return nil
	},
}

var solCmd = &cobra.Command{
	Use:   "sol",
	Short: "Ouvrir la console série (Serial over LAN) — ~. pour quitter",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := target()
		if err != nil {
			return err
		}
		c := t.SOL()
		c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
		return c.Run()
	},
}

func infoCmd(use, short string, f func(ipmi.Target) (string, error)) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := target()
			if err != nil {
				return err
			}
			return output(f(t))
		},
	}
}

var rawCmd = &cobra.Command{
	Use:   "raw -- <arguments ipmitool>",
	Short: "Passer des arguments directement à ipmitool",
	Example: "  ipmi-menu raw -s srv1 -- chassis status\n" +
		"  ipmi-menu raw -s srv1 -- user list 1",
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := target()
		if err != nil {
			return err
		}
		return output(t.Run(ipmi.Slow, args...))
	},
}

func init() {
	bootCmd.Flags().BoolVar(&bootLegacy, "legacy", false, "démarrage Legacy/BIOS au lieu d'UEFI")
	bootCmd.Flags().BoolVar(&bootPersistent, "persistent", false, "pour tous les démarrages, pas seulement le prochain")
	bootCmd.Flags().BoolVar(&bootReboot, "reboot", false, "redémarrer (ou allumer) le serveur ensuite")

	rootCmd.AddCommand(powerCmd, bootCmd, solCmd, rawCmd,
		infoCmd("sensors", "Afficher les capteurs (températures, ventilateurs, alimentations…)", ipmi.Target.Sensors),
		infoCmd("hw", "Afficher les infos matériel (BMC + FRU)", ipmi.Target.Hardware),
		infoCmd("lan", "Afficher la configuration réseau du BMC", ipmi.Target.LAN),
	)
}
