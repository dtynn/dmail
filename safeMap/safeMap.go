package safeMap

import (
	"fmt"
	"time"
)

const (
	actionSet int = iota
	actionSetex
	actionGet
	actionDel
	actionSetnx
)

var (
	nonDead int64 = -1
	nonData       = &data{}

	errCacheNotFound = fmt.Errorf("cache not found")
	errCacheExists   = fmt.Errorf("cache exists")
)

type action struct {
	action int
	key    interface{}
	data   *data
	ch     chan *data
	chErr  chan error
}

type data struct {
	val  interface{}
	dead int64
}

func isNonData(d *data) bool {
	return d == nonData
}

type SafeMap struct {
	m       map[interface{}]*data
	actions chan *action
}

func NewSafeMap() *SafeMap {
	s := &SafeMap{
		m:       map[interface{}]*data{},
		actions: make(chan *action),
	}
	go func(s *SafeMap) {
		s.run()
	}(s)
	return s
}

func (this *SafeMap) Set(key, val interface{}) error {
	a := &action{
		action: actionSet,
		key:    key,
		data: &data{
			val:  val,
			dead: nonDead,
		},
	}
	this.actions <- a
	return nil
}

func (this *SafeMap) Setex(key, val interface{}, expire int64) error {
	dead := time.Now().Add(time.Duration(expire) * time.Second).Unix()
	a := &action{
		action: actionSetex,
		key:    key,
		data: &data{
			val:  val,
			dead: dead,
		},
	}
	this.actions <- a
	return nil
}

func (this *SafeMap) Setnx(key, val interface{}) error {
	ch := make(chan error, 1)
	a := &action{
		action: actionSetnx,
		key:    key,
		data: &data{
			val:  val,
			dead: nonDead,
		},
		chErr: ch,
	}
	this.actions <- a
	return <-ch
}

func (this *SafeMap) Get(key interface{}) (interface{}, error) {
	ch := make(chan *data)
	a := &action{
		action: actionGet,
		key:    key,
		ch:     ch,
	}
	this.actions <- a

	d := <-ch

	if isNonData(d) {
		return nil, errCacheNotFound
	}

	now := time.Now().Unix()
	if d.dead != nonDead && d.dead < now {
		this.Del(key)
		return nil, errCacheNotFound
	}

	return d.val, nil
}

func (this *SafeMap) Del(key interface{}) error {
	a := &action{
		action: actionDel,
		key:    key,
	}
	this.actions <- a
	return nil
}

func (this *SafeMap) run() {
	for a := range this.actions {
		switch a.action {
		case actionSet, actionSetex:
			this.m[a.key] = a.data
		case actionGet:
			if d, has := this.m[a.key]; has {
				a.ch <- d
			} else {
				a.ch <- nonData
			}
		case actionDel:
			delete(this.m, a.key)
		case actionSetnx:
			if _, has := this.m[a.key]; has {
				a.chErr <- errCacheExists
			} else {
				this.m[a.key] = a.data
				a.chErr <- nil
			}
		}
	}
}
