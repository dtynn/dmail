package dmail

import (
	"time"
)

type header interface {
	Field() string
	Value() string
	String() string
}

type normalHeader struct {
	field, value string
}

func (this *normalHeader) Field() string {
	return this.field
}

func (this *normalHeader) Value() string {
	return this.value
}

func (this *normalHeader) String() string {
	return this.field + ": " + this.value
}

type addressHeader struct {
	field, address, name string
}

func (this *addressHeader) Field() string {
	return this.field
}

func (this *addressHeader) Value() string {
	return this.name + "<" + this.address + ">"
}

func (this *addressHeader) String() string {
	return this.field + ": " + this.Value()
}

type dateHeader struct {
	field string
	date  time.Time
}

func (this *dateHeader) Field() string {
	return this.field
}

func (this *dateHeader) Value() string {
	return this.date.Format(time.RFC822Z)
}

func (this *dateHeader) String() string {
	return this.field + ": " + this.Value()
}
