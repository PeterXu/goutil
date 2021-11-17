package util

import (
	"sort"
	"sync"
)

/// List

type fnListLess func(a, b interface{}) bool

type List struct {
	Queue
	less fnListLess
}

func NewList(less fnListLess) *List {
	return &List{
		less: less,
	}
}

func (t *List) Add(e interface{}) {
	t.Lock()
	defer t.Unlock()

	t.l = append(t.l, e)
	if t.less != nil {
		max_idx := len(t.l) - 1
		for idx := max_idx - 1; idx >= 0; idx -= 1 {
			item := t.l[idx]
			if t.less(item, e) { // stable sort
				break
			} else {
				t.l[idx+1] = item
				t.l[idx] = e
			}
		}
	}
}

func (t *List) Sort(less fnListLess) {
	t.Lock()
	defer t.Unlock()

	if less != nil {
		old := t.less
		t.less = less
		sort.Sort(t)
		t.less = old
	}
}

func (t *List) SortStable(less fnListLess) {
	t.Lock()
	defer t.Unlock()

	if less != nil {
		old := t.less
		t.less = less
		sort.Stable(t)
		t.less = old
	}
}

// @sort only and no mutex here
func (t *List) Less(i, j int) bool {
	if t.less != nil {
		return t.less(t.l[i], t.l[j])
	} else {
		return true
	}
}

// @sort only and no mutex here
func (t *List) Swap(i, j int) {
	tmp := t.l[i]
	t.l[i] = t.l[j]
	t.l[j] = tmp
}

/// Queue

type Queue struct {
	sync.Mutex
	l []interface{}
}

func NewQueue() *Queue {
	return &Queue{}
}

func (q *Queue) Len() int {
	q.Lock()
	defer q.Unlock()

	return len(q.l)
}

func (q *Queue) Get(idx int) interface{} {
	q.Lock()
	defer q.Unlock()

	if idx >= 0 && idx < len(q.l) {
		return q.l[idx]
	}
	return nil
}

func (q *Queue) Empty() bool {
	q.Lock()
	defer q.Unlock()

	return len(q.l) == 0
}

func (q *Queue) Clear() {
	q.Lock()
	defer q.Unlock()

	q.l = q.l[0:0]
}

// Add a new element to the end, after its current last element.
func (q *Queue) PushBack(e interface{}) {
	q.Lock()
	defer q.Unlock()

	q.l = append(q.l, e)
}

// Pop the last element
func (q *Queue) PopBack() {
	q.Lock()
	defer q.Unlock()

	if len(q.l) != 0 {
		size := len(q.l) - 1
		q.l = q.l[:size]
	}
}

// Access the last element
func (q *Queue) Back() interface{} {
	q.Lock()
	defer q.Unlock()

	if len(q.l) != 0 {
		idx := len(q.l) - 1
		return q.l[idx]
	}
	return nil
}

// Insert a new element to the front, before its current first element.
func (q *Queue) PushFront(e interface{}) {
	q.Lock()
	defer q.Unlock()

	var tmpl []interface{}
	tmpl = append(tmpl, e)
	q.l = append(tmpl, q.l...)
}

// Pop the first element
func (q *Queue) PopFront() {
	q.Lock()
	defer q.Unlock()

	if len(q.l) != 0 {
		q.l = q.l[1:]
	}
}

// Access the first element
func (q *Queue) Front() interface{} {
	q.Lock()
	defer q.Unlock()

	if len(q.l) != 0 {
		return q.l[0]
	}
	return nil
}

// Access values @unsafe
func (q *Queue) Values() []interface{} {
	return q.l
}
