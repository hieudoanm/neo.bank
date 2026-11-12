package db

import (
	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

func NewSession(hosts []string, keyspace string) (*gocql.Session, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.LocalOne
	return cluster.CreateSession()
}
