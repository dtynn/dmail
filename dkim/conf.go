package dkim

type DkimConf struct {
	domain, identity, selector string
	setLength                  bool
	pemBytes                   []byte
}

func NewDkimConf(domain, identity, selector string, setLength bool, pemBytes []byte) *DkimConf {
	return &DkimConf{domain, identity, selector, setLength, pemBytes}
}
