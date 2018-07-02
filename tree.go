package browscap

import (
	"strings"
)

type radixTree struct {
	leafs radix
	value *Browser
}

type radix map[uint8]*radixTree

const (
	skipOne      uint8 = '?'
	skipInfinite       = '*'
)

func (r *radixTree) Add(opts []string, br *Browser) {
	br.mapArray(opts)
	r.add(strings.ToLower(fPropertyName.GetString(opts)), br)
}

func (r *radixTree) add(search string, br *Browser) {
	if search == "*" || search == "" {
		r.value = br

		return
	} else if r.leafs == nil {
		r.leafs = make(radix)
	}

	c := search[0]

	if v, ok := r.leafs[c]; ok {
		v.add(search[1:], br)
	} else {
		rt := new(radixTree)
		rt.add(search[1:], br)
		r.leafs[c] = rt
	}
	r.leafs[c] = r.leafs[c]
}

func (r *radixTree) Find(search string) *Browser {
	if f := r.find(strings.ToLower(search)); len(f) > 0 {
		return f[0]
	}

	return nil
}

func (r *radixTree) find(search string) (found []*Browser) {
	if len(search) == 0 {
		return
	} else if r.leafs == nil {
		found = append(found, r.value)
	}

	current := search[0]

	if leaf, ok := r.leafs[current]; ok {
		found = append(found, leaf.find(search[1:])...)
	}

	if leaf, ok := r.leafs[skipOne]; ok {
		found = append(found, leaf.find(search[1:])...)
	}

	if leaf, ok := r.leafs[skipInfinite]; ok {
		for i := 0; i < len(search); i++ {
			current := search[i]
			if leaf, ok := leaf.leafs[current]; ok {
				found = append(found, leaf.find(search[i+1:])...)
			}
		}
	}

	return
}

func (r radixTree) String() string {
	if len(r.leafs) != 0 {
		str := "{"
		for i := range r.leafs {
			str += string(i) + ","
		}

		if l := len(str); l > 1 {
			return str[:l-1] + "}"
		}
	} else if r.value != nil {
		return r.value.String()
	}

	return ""
}