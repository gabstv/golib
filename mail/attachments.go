package mail

import (
	"io"
)

type AttachmentInfo struct {
	File     io.ReadCloser
	Name     string
	MimeType string
}

func NewAttachmentList() *AttachmentList {
	v := &AttachmentList{}
	v.count = 0
	return v
}

type AttachmentList struct {
	count int
	first *AttachmentListItem
	last  *AttachmentListItem
}

func (l *AttachmentList) Add(item *AttachmentInfo) {
	ni := &AttachmentListItem{}
	ni.Value = item
	if l.first == nil {
		l.first = ni
	}
	if l.last != nil {
		l.last.next = ni
		ni.prev = l.last
	}
	l.last = ni
	l.count++
}

func (l *AttachmentList) First() *AttachmentListItem {
	return l.first
}

func (l *AttachmentList) Last() *AttachmentListItem {
	return l.last
}

func (l *AttachmentList) Count() int {
	return l.count
}

func (l *AttachmentList) GetFilenames() []string {
	//P141006 fixed memory leak
	names := make([]string, 0)
	for li := l.First(); li != nil; li = li.Next() {
		names = append(names, li.Value.Name)
	}
	return names
}

type AttachmentListItem struct {
	Value *AttachmentInfo
	next  *AttachmentListItem
	prev  *AttachmentListItem
}

func (q *AttachmentListItem) Next() *AttachmentListItem {
	return q.next
}
