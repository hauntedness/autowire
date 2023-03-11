package util

import "fmt"

type BiMap[L, R comparable] struct {
	lMap map[L]R
	rMap map[R]L
}

func NewBiMap[L, R comparable]() BiMap[L, R] {
	return BiMap[L, R]{lMap: make(map[L]R), rMap: make(map[R]L)}
}

// # Put l and r into map,
//
//	put report false if either l or r is already mapped to other value
func (bm BiMap[L, R]) Put(l L, r R) (ok bool) {
	if rv, ok := bm.lMap[l]; ok && rv != r {
		return false
	}
	if lv, ok := bm.rMap[r]; ok && lv != l {
		return false
	}
	bm.lMap[l] = r
	bm.rMap[r] = l
	return true
}

// # Put l and r into map,
//
//	MustPut panic if either l or r is already mapped to other value
func (bm BiMap[L, R]) MustPut(l L, r R) {
	if rv, ok := bm.lMap[l]; ok && rv != r {
		panic(fmt.Errorf("lv %v already mapped to %v", l, rv))
	}
	if lv, ok := bm.rMap[r]; ok && lv != l {
		panic(fmt.Errorf("rv %v already mapped to %v", r, lv))
	}
	bm.lMap[l] = r
	bm.rMap[r] = l
}

func (bm BiMap[L, R]) GetByL(key L) (rv R, ok bool) {
	r, ok := bm.lMap[key]
	return r, ok
}

func (bm BiMap[L, R]) GetByR(r R) (lv L, ok bool) {
	l, ok := bm.rMap[r]
	return l, ok
}

func (bm BiMap[L, R]) LMap() map[L]R {
	return bm.lMap
}
