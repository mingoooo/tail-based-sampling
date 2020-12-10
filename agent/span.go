package agent

import (
	"strings"

	pb "github.com/mingoooo/tail-based-sampling/g"
	"github.com/mingoooo/tail-based-sampling/utils"
)

func NewSpan(line []byte) *Span {
	s := &Span{}
	s.Raw = utils.ByteSliceToString(line)
	tidEndIndex := strings.IndexByte(s.Raw, '|')
	s.TraceID = s.Raw[:tidEndIndex]
	s.filterWrongSpan()
	return s
}

func (s *Span) filterWrongSpan() {
	tagFirstIndex := strings.LastIndexByte(s.Raw, '|')
	tags := s.Raw[tagFirstIndex:]
	if strings.Index(tags, "error=1") != -1 {
		s.Wrong = true
	}
	if strings.Index(tags, "http.status_code=") != -1 && strings.Index(tags, "http.status_code=200") == -1 {
		s.Wrong = true
	}
}

func ParseSpan(s string) *pb.Span {
	firstIndex := strings.IndexByte(s, '|')
	secondIndex := strings.IndexByte(s[firstIndex+1:], '|')
	return &pb.Span{
		Raw:       s,
		StartTime: s[firstIndex+1 : firstIndex+1+secondIndex],
	}
}