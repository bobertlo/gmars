package mars

import "fmt"

type processQueue struct {
	queue  []Address
	size   Address
	length Address
	start  Address
	end    Address
}

func NewProcessQueue(size Address) *processQueue {
	queue := make([]Address, size)

	return &processQueue{
		queue: queue,
		size:  size,
	}
}

func (q *processQueue) Len() Address {
	return q.length
}

func (q *processQueue) Push(a Address) {
	if q.length >= q.size {
		return
	}
	q.queue[q.end] = a
	q.end = (q.end + 1) % q.size
	q.length++
}

func (q *processQueue) Pop() (Address, error) {
	if q.length == 0 {
		return 0, fmt.Errorf("pull from empty queue")
	}
	val := q.queue[q.start]
	q.start++
	q.length--
	return val, nil
}

func (q *processQueue) get(n Address) Address {
	return q.queue[(q.start+n)%q.size]
}

func (q *processQueue) Values() []Address {
	dat := make([]Address, q.Len())
	for i := Address(0); i < q.Len(); i++ {
		dat[i] = q.get(i)
	}
	return dat
}
