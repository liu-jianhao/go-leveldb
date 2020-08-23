package skiplist

type Comparator func(a, b interface{}) int

func IntComparator(a, b interface{}) int {
	aInt, bInt := a.(int), b.(int)
	return aInt - bInt
}
