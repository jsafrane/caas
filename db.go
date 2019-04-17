package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/gocql/gocql"
)

type DB interface {
	IncrementAndGet(counterName string) (Counter, error)
}

type cassandraDB struct {
	session  *gocql.Session
	hostname string
}

func NewCassandra() (DB, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	addr := os.Getenv("CASSANDRA_ADDRESS")
	if addr == "" {
		return nil, fmt.Errorf("CASSANDRA_ADDRESS must be set")
	}

	// We need to resolve all IPs of cassandra server and connect to them
	ips, err := net.LookupHost(addr)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve %s: %s", addr, err)
	}

	log.Printf("Resolved cassandra address %s to %+v", addr, ips)

	cluster := gocql.NewCluster(ips...)
	cluster.Keyspace = "caas"

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	return &cassandraDB{session, hostname}, nil
}

func (c *cassandraDB) IncrementAndGet(counterName string) (Counter, error) {
	observer := &queryLogger{}
	err := c.increment(counterName, observer)
	if err != nil {
		return Counter{}, fmt.Errorf("error incrementing %q: %s", err)
	}

	count, err := c.get(counterName, observer)
	if err != nil {
		return Counter{}, fmt.Errorf("error getting %q: %s", err)
	}

	return Counter{Value: count, Name: counterName, Host: c.hostname, DBStats: observer.stats}, nil
}

func (c *cassandraDB) increment(name string, observer gocql.QueryObserver) error {
	query := c.session.Query(`UPDATE counter SET value=value+1 WHERE name = ?`, name)
	query.Observer(observer)
	return query.Exec()
}

func (c *cassandraDB) get(name string, observer gocql.QueryObserver) (count int64, err error) {
	m := map[string]interface{}{}
	cql := "SELECT name, value FROM counter WHERE name=? LIMIT 1"
	query := c.session.Query(cql, name).Consistency(gocql.One)
	query.Observer(observer)
	if err := query.MapScan(m); err != nil {
		return 0, err
	}
	return m["value"].(int64), nil
}

type queryLogger struct {
	stats []QueryStat
}

var _ gocql.QueryObserver = &queryLogger{}

func (q *queryLogger) ObserveQuery(_ context.Context, query gocql.ObservedQuery) {
	if q.stats == nil {
		q.stats = []QueryStat{}
	}
	stat := QueryStat{
		Statement: query.Statement,
		Attempts:  query.Metrics.Attempts,
		Time:      fmt.Sprintf("%f miliseconds", query.End.Sub(query.Start).Seconds()*1000),
		Host:      query.Host.ConnectAddress().String(),
		Rows:      query.Rows,
	}
	q.stats = append(q.stats, stat)
}
