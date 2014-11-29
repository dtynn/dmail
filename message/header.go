package message

import (
	"time"
)

type Header interface {
	Field() string
	Value() string
	String() string
}

type NormalHeader struct {
	field, value string
}

func NewNormalHeader(field, value string) *NormalHeader {
	return &NormalHeader{field, value}
}

func (this *NormalHeader) Field() string {
	return this.field
}

func (this *NormalHeader) Value() string {
	return this.value
}

func (this *NormalHeader) String() string {
	return this.field + ": " + this.value
}

type AddressHeader struct {
	field, address, name string
}

func NewAddressHeader(field, address, name string) *AddressHeader {
	return &AddressHeader{field, address, name}
}

func (this *AddressHeader) Field() string {
	return this.field
}

func (this *AddressHeader) Value() string {
	return this.name + "<" + this.address + ">"
}

func (this *AddressHeader) String() string {
	return this.field + ": " + this.Value()
}

type DateHeader struct {
	field string
	date  time.Time
}

func NewDateHeader(field string, date time.Time) *DateHeader {
	return &DateHeader{field, date}
}

func (this *DateHeader) Field() string {
	return this.field
}

func (this *DateHeader) Value() string {
	return this.date.Format(time.RFC822Z)
}

func (this *DateHeader) String() string {
	return this.field + ": " + this.Value()
}
