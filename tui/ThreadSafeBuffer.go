package tui

import (
	"bytes"
	"sync"
)

type TBuffer struct {
    b bytes.Buffer
    m sync.Mutex
}
func (b *TBuffer) Read(p []byte) (n int, err error) {
    b.m.Lock()
    defer b.m.Unlock()
    return b.b.Read(p)
}
func (b *TBuffer) Write(p []byte) (n int, err error) {
    b.m.Lock()
    defer b.m.Unlock()
    return b.b.Write(p)
}
func (b *TBuffer) String() string {
    b.m.Lock()
    defer b.m.Unlock()
    return b.b.String()
}
func (b *TBuffer) WriteString(s string) (n int, err error) {
    b.m.Lock()
    defer b.m.Unlock()
    return b.b.WriteString(s)
}
func (b *TBuffer) Len() int {
    b.m.Lock()
    defer b.m.Unlock()
    return b.b.Len()
}