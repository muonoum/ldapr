package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/go-ldap/ldap/v3"
	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Server   string
	User     string
	Password string
	Base     string
}

func configure(path string) (Config, error) {
	var config Config
	f, err := os.Open(path)
	if err != nil {
		return config, err
	}
	err = toml.NewDecoder(f).Decode(&config)
	return config, err
}

func main() {
	flag.Usage = func() {
		fmt.Println(`ldapr [-config PATH] -filter "TERM" [-attributes NAME,...] -template "TEMPLATE"`)
		os.Exit(2)
	}

	term := flag.String("term", "", "")
	attr := flag.String("attr", "", "")
	tmpl := flag.String("tmpl", "", "")
	cf := flag.String("config", "/etc/ldapr.toml", "")

	flag.Parse()

	config, err := configure(*cf)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var conn *ldap.Conn

	if strings.HasSuffix(config.Server, ":636") {
		if conn, err = ldap.DialTLS("tcp", config.Server, &tls.Config{}); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		if conn, err = ldap.Dial("tcp", config.Server); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if err := conn.Bind(config.User, config.Password); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var attrs []string
	if *attr != "" {
		attrs = strings.Split(*attr, ",")
	}

	if res, err := conn.SearchWithPaging(ldap.NewSearchRequest(config.Base,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		*term, attrs, []ldap.Control{},
	), 1000); err != nil {
		fmt.Println(err)
		os.Exit(1)
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
						fmt.Println(err)
						os.Exit(1)
					}
					if err := tmp.Execute(os.Stdout, m); err != nil {
						fmt.Println(err)
						os.Exit(1)
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
