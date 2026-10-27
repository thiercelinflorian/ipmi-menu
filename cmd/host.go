package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/thiercelinflorian/ipmi-menu/internal/config"
	"github.com/thiercelinflorian/ipmi-menu/internal/ipmi"
	"golang.org/x/term"
)

var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "Gérer les serveurs enregistrés",
	Long:  "Les serveurs sont stockés dans " + config.Path() + " (fichier en 0600).",
}

var hostListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Lister les serveurs",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NOM\tADRESSE\tUTILISATEUR")
		for _, name := range cfg.Names() {
			t, _ := cfg.Get(name)
			user := t.User
			if user == "" {
				user = ipmi.DefaultUser
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", name, t.Host, user)
		}
		return w.Flush()
	},
}

var hostAddCmd = &cobra.Command{
	Use:     "add <nom> <adresse>",
	Short:   "Ajouter ou remplacer un serveur",
	Example: "  ipmi-menu host add srv1 10.0.0.10 -U ADMIN",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		t := ipmi.Target{Name: args[0], Host: args[1], User: flagUser, Port: flagPort, Password: passwordFlag()}
		if t.Password == "" && term.IsTerminal(int(os.Stdin.Fd())) {
			fmt.Fprint(os.Stderr, "Mot de passe (vide = aucun) : ")
			pw, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Fprintln(os.Stderr)
			if err != nil {
				return err
			}
			t.Password = string(pw)
		}
		cfg.Put(t)
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Println(t.Name, "enregistré")
		return nil
	},
}

var hostRmCmd = &cobra.Command{
	Use:     "rm <nom>",
	Aliases: []string{"remove"},
	Short:   "Supprimer un serveur",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		cfg, _ := config.Load()
		if cfg == nil {
			return nil, cobra.ShellCompDirectiveError
		}
		return cfg.Names(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if _, ok := cfg.Get(args[0]); !ok {
			return fmt.Errorf("serveur %q inconnu", args[0])
		}
		cfg.Delete(args[0])
		return cfg.Save()
	},
}

func init() {
	hostCmd.AddCommand(hostListCmd, hostAddCmd, hostRmCmd)
	rootCmd.AddCommand(hostCmd)
}
