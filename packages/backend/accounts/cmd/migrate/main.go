package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

func main() {
	cp := "localhost:9042"
	if len(os.Args) > 1 {
		cp = os.Args[1]
	}
	cluster := gocql.NewCluster(strings.Split(cp, ",")...)
	cluster.Keyspace = "account_keyspace"
	cluster.Timeout = 10 * time.Second
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer session.Close()
	data, err := os.ReadFile("internal/db/migrations/001_create_tables.cql")
	if err != nil {
		log.Fatalf("read mig: %v", err)
	}
	for _, st := range splitStmts(string(data)) {
		st = strings.TrimSpace(st)
		if st == "" || strings.HasPrefix(st, "--") {
			continue
		}
		if err := session.Query(st).Exec(); err != nil {
			log.Fatalf("exec %s: %v", trunc(st, 60), err)
		}
	}
	log.Println("migrations applied")
}
func splitStmts(cql string) []string {
	var out []string
	var b strings.Builder
	ins := false
	for i := 0; i < len(cql); i++ {
		ch := cql[i]
		if ins {
			b.WriteByte(ch)
			if ch == '\'' && i+1 < len(cql) && cql[i+1] == '\'' {
				b.WriteByte(cql[i+1])
				i++
			} else if ch == '\'' {
				ins = false
			}
			continue
		}
		if ch == '\'' {
			ins = true
			b.WriteByte(ch)
			continue
		}
		if ch == ';' {
			out = append(out, b.String())
			b.Reset()
			continue
		}
		b.WriteByte(ch)
	}
	if b.Len() > 0 {
		out = append(out, b.String())
	}
	return out
}
func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

var _ = fmt.Sprintf
