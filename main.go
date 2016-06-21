package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/go-ldap/ldap"
	"github.com/spektroskop/util"
	"github.com/spf13/viper"
)

func main() {
	flag.Usage = func() {
		fmt.Println(`ldapr -term "TERM" [-attr NAME,...] -tmpl "TEMPLATE"`)
		os.Exit(2)
	}

	term := flag.String("term", "", "")
	attr := flag.String("attr", "", "")
	tmpl := flag.String("tmpl", "", "")

	flag.Parse()

	viper.SetConfigName("ldapr")
	viper.SetConfigType("toml")

	viper.AddConfigPath(".")
	viper.AddConfigPath("~")

	if err := viper.ReadInConfig(); err != nil {
		util.Error(err, 1)
	}

	conn, err := ldap.Dial("tcp", viper.GetString("server"))
	if err != nil {
		util.Error(err, 1)
	}

	if err := conn.Bind(viper.GetString("user"), viper.GetString("password")); err != nil {
		util.Error(err, 1)
	}

	var attrs []string
	if *attr != "" {
		attrs = strings.Split(*attr, ",")
	}

	if res, err := conn.SearchWithPaging(ldap.NewSearchRequest(viper.GetString("base"),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		*term, attrs, []ldap.Control{},
	), 1000); err != nil {
		util.Error(err, 1)
	} else {
		if *tmpl != "" {
			for _, entry := range res.Entries {
				m := map[string]string{
					"DN": entry.DN,
				}

				for _, attr := range entry.Attributes {
					m[attr.Name] = strings.Join(attr.Values, ",")
				}

				if *tmpl != "" {
					tmp, err := template.New("ldapr").Parse(strings.TrimSpace(*tmpl) + "\n")
					if err != nil {
						util.Error(err, 1)
					}
					if err := tmp.Execute(os.Stdout, m); err != nil {
						util.Error(err, 1)
					}
				}
			}
		} else {
			for _, entry := range res.Entries {
				entry.PrettyPrint(0)
			}
		}

	}
}
