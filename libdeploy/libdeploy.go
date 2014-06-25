package libdeploy

type node struct {
	value interface{}
	next  *node
}

func NewStack() *Stack {
	return &Stack{}
}

type Stack struct {
	head  *node
	count int
}

func (s *Stack) Push(val interface{}) {
	n := &node{value: val}

	if s.head == nil {
		s.head = n
	} else {
		n.next = s.head
		s.head = n
	}

	s.count++
}

func (s *Stack) Pop() interface{} {
	var n *node
	if s.head != nil {
		n = s.head
		s.head = n.next
		s.count--
	}

	if n == nil {
		return nil
	}

	return n.value
}

func (s *Stack) Len() int {
	return s.count
}
