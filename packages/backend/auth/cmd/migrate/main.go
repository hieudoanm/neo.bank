package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gocql/gocql"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate <cassandra-contact-points>")
	}
	contactPoints := strings.Split(os.Args[1], ",")

	cluster := gocql.NewCluster(contactPoints...)
	cluster.Keyspace = "auth_keyspace"
	cluster.Timeout = 10 * time.Second
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("connect cassandra: %v", err)
	}
	defer session.Close()

	if err := applyMigration(session, "internal/db/migrations/001_create_tables.cql"); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("migrations applied successfully")
}

func applyMigration(session *gocql.Session, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read migration: %w", err)
	}
	for _, stmt := range splitStatements(string(data)) {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		if err := session.Query(stmt).Exec(); err != nil {
			return fmt.Errorf("execute statement %q: %w", truncate(stmt, 80), err)
		}
	}
	return nil
}

func splitStatements(cql string) []string {
	var out []string
	var current strings.Builder
	inString := false
	for i := 0; i < len(cql); i++ {
		ch := cql[i]
		if inString {
			current.WriteByte(ch)
			if ch == '\'' && i+1 < len(cql) && cql[i+1] == '\'' {
				current.WriteByte(cql[i+1])
				i++
			} else if ch == '\'' {
				inString = false
			}
			continue
		}
		if ch == '\'' {
			inString = true
			current.WriteByte(ch)
			continue
		}
		if ch == ';' {
			out = append(out, current.String())
			current.Reset()
			continue
		}
		current.WriteByte(ch)
	}
	if current.Len() > 0 {
		out = append(out, current.String())
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
