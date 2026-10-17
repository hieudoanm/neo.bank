package main

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

const defaultHosts = "localhost"

var errNoKeyspace = errors.New("no CREATE KEYSPACE statement found in migration")

func main() {
	hosts := strings.Split(env("CASSANDRA_HOST", defaultHosts), ",")

	data, err := os.ReadFile("internal/db/migrations/001_create_tables.cql")
	if err != nil {
		log.Fatalf("read migration: %v", err)
	}
	stmts := splitStmts(string(data))

	keyspace, err := parseKeyspace(stmts)
	if err != nil {
		log.Fatalf("parse keyspace: %v", err)
	}

	if err := createKeyspace(hosts, stmts); err != nil {
		log.Fatalf("create keyspace: %v", err)
	}
	if err := applyMigrations(hosts, keyspace, stmts); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}
	log.Printf("migrations applied to %s", keyspace)
}

func parseKeyspace(stmts []string) (string, error) {
	for _, st := range stmts {
		if strings.HasPrefix(upper(st), "CREATE KEYSPACE") {
			fields := strings.Fields(st)
			for i, f := range fields {
				if strings.EqualFold(f, "EXISTS") && i+1 < len(fields) {
					return fields[i+1], nil
				}
			}
			if len(fields) >= 3 {
				return fields[2], nil
			}
		}
	}
	return "", errNoKeyspace
}

func createKeyspace(hosts []string, stmts []string) error {
	session, err := newSession(hosts, "")
	if err != nil {
		return err
	}
	defer session.Close()
	for _, st := range stmts {
		if strings.HasPrefix(upper(st), "CREATE KEYSPACE") {
			if err := session.Query(st).Exec(); err != nil {
				return err
			}
		}
	}
	return nil
}

func applyMigrations(hosts []string, keyspace string, stmts []string) error {
	session, err := newSession(hosts, keyspace)
	if err != nil {
		return err
	}
	defer session.Close()
	for _, st := range stmts {
		up := upper(st)
		if strings.HasPrefix(up, "USE") || strings.HasPrefix(up, "CREATE KEYSPACE") {
			continue
		}
		if err := session.Query(st).Exec(); err != nil {
			return err
		}
	}
	return nil
}

func newSession(hosts []string, keyspace string) (*gocql.Session, error) {
	cluster := gocql.NewCluster(hosts...)
	if keyspace != "" {
		cluster.Keyspace = keyspace
	}
	cluster.Timeout = 10 * time.Second
	return cluster.CreateSession()
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func upper(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
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
			if st := strings.TrimSpace(b.String()); st != "" {
				out = append(out, st)
			}
			b.Reset()
			continue
		}
		b.WriteByte(ch)
	}
	if st := strings.TrimSpace(b.String()); st != "" {
		out = append(out, st)
	}
	return out
}
