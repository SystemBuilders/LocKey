package cache

import (
	"fmt"
	"testing"
)

func Test_LRUCache(t *testing.T) {
	lruCache := NewLRUCache(5)

	one := NewSimpleKey(1)
	two := NewSimpleKey(2)
	three := NewSimpleKey(3)
	four := NewSimpleKey(4)

	err := lruCache.PutElement(one)
	if err != nil {
		fmt.Println(err)
	}
	lruCache.PrintCache()

	err = lruCache.PutElement(two)
	if err != nil {

	}
	lruCache.PrintCache()

	err = lruCache.GetElement(one)
	if err != nil {
		fmt.Println(err)
	}
	lruCache.PrintCache()

	err = lruCache.PutElement(three)
	if err != nil {
		fmt.Println(err)
	}
	lruCache.PrintCache()

	err = lruCache.GetElement(two)
	if err != nil {
		fmt.Println(err)
	}
	lruCache.PrintCache()

	err = lruCache.PutElement(four)
	if err != nil {
		fmt.Println(err)
	}
	lruCache.PrintCache()

	err = lruCache.GetElement(one)
	if err != nil {
		fmt.Println(err)
	}
	lruCache.PrintCache()

	err = lruCache.GetElement(three)
	if err != nil {
		fmt.Println(err)
	}
	lruCache.PrintCache()

	err = lruCache.GetElement(four)
	if err != nil {
		fmt.Println(err)
	}
	lruCache.PrintCache()
}
