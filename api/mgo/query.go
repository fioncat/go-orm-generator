package mgo

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Query struct {
	mq   *mgo.Query
	sess *mgo.Session
}

func NewQuery(mq *mgo.Query, sess *mgo.Session) *Query {
	return &Query{mq: mq, sess: sess}
}

func (q *Query) Select(fields ...string) *Query {
	m := make(bson.M, len(fields))
	for _, field := range fields {
		m[field] = 1
	}
	q.mq.Select(m)
	return q
}

func (q *Query) Limit(offset, limit int) *Query {
	if limit > 0 {
		q.mq.Limit(limit)
	}
	if offset > 0 {
		q.mq.Skip(offset)
	}
	return q
}

func (q *Query) Sort(fields ...string) *Query {
	q.mq.Sort(fields...)
	return q
}

func (q *Query) Batch(n int) *Query {
	q.mq.Batch(n)
	return q
}

func (q *Query) Prefetch(p float64) *Query {
	q.mq.Prefetch(p)
	return q
}

func (q *Query) Iter() *mgo.Iter {
	return q.mq.Iter()
}

func (q *Query) MarshalAll(v interface{}) error {
	defer q.sess.Close()
	return q.mq.All(v)
}

func (q *Query) MarshalOne(v interface{}) error {
	defer q.sess.Close()
	return q.mq.One(v)
}
