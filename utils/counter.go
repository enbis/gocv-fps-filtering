package utils

type Counter struct {
	count int
}

func NewCounter(val int) *Counter {

	c := &Counter{
		count: val,
	}
	return c
}

func (c *Counter) SetCounter(val int) {
	c.count = val
}

func (c *Counter) Decrement() {
	c.count -= 1
}

func (c *Counter) Increment() {
	c.count += 1
}

func (c *Counter) GetCount() int {
	return c.count
}

func (c *Counter) GetInitVal() int {
	return c.count
}
