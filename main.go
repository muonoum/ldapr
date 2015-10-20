package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/go-ldap/ldap"
	"github.com/spf13/viper"
)

func main() {
	flag.Usage = func() {
		fmt.Println("ldapr <search> [attributes]")
		os.Exit(2)
	}

	flag.Parse()

	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	conn, err := ldap.Dial("tcp", viper.GetString("server"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := conn.Bind(viper.GetString("user"), viper.GetString("password")); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
	}

	search := flag.Args()[0]
	var attrs []string
	if len(args) > 1 {
		attrs = strings.Split(args[1], ",")
	}

	if res, err := conn.SearchWithPaging(ldap.NewSearchRequest(viper.GetString("base"),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		search, attrs, []ldap.Control{},
	), 1000); err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		for _, entry := range res.Entries {
			entry.PrettyPrint(0)
		}
	}
}
