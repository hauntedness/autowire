package util

import "fmt"

type BiDirectionMap[L, R comparable] struct {
	lMap map[L]R
	rMap map[R]L
}

func NewBiDirectionMap[L, R comparable]() BiDirectionMap[L, R] {
	return BiDirectionMap[L, R]{lMap: make(map[L]R), rMap: make(map[R]L)}
}

// # Put l and r into map,
//
//	put report false if either l or r is already stored
func (bm BiDirectionMap[L, R]) Put(l L, r R) (ok bool) {
	if _, ok := bm.lMap[l]; ok {
		return false
	}
	if _, ok := bm.rMap[r]; ok {
		return false
	}
	bm.lMap[l] = r
	bm.rMap[r] = l
	return true
}

// # Put l and r into map,
//
//	put report false if either l or r is already stored
func (bm BiDirectionMap[L, R]) MustPut(l L, r R) {
	if rv, ok := bm.lMap[l]; ok {
		panic(fmt.Errorf("left side value already mapped to %v", rv))
	}
	if lv, ok := bm.rMap[r]; ok {
		panic(fmt.Errorf("right side value already mapped to %v", lv))
	}
	bm.lMap[l] = r
	bm.rMap[r] = l
}

func (bm BiDirectionMap[L, R]) GetByL(key L) (rv R, ok bool) {
	r, ok := bm.lMap[key]
	return r, ok
}

func (bm BiDirectionMap[L, R]) GetByR(r R) (lv L, ok bool) {
	l, ok := bm.rMap[r]
	return l, ok
}
