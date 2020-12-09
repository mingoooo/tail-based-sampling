package agent

type Ringbuf struct {
	size int
	buf  []byte
	used int
}

func NewRingbuf(size int) *Ringbuf {
	return &Ringbuf{
		size: size,
		buf:  make([]byte, size),
	}
}

func (r *Ringbuf) Write(data []byte) {

}
