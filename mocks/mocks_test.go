package mocks_test

import (
	"bytes"
	"go/ast"
	"go/format"
	"testing"

	"github.com/a8m/expect"
	"github.com/nelsam/hel/mocks"
)

func TestGenerate(t *testing.T) {
	expect := expect.New(t)

	types := []*ast.TypeSpec{
		typeSpec(expect, `
  type Foo interface {
   Bar() int
  }`),
		typeSpec(expect, `
  type Bar interface {
   Foo(foo string)
   Baz()
  }`),
	}

	mockFinder := newMockTypeFinder()
	mockFinder.ExportedTypesOutput.ret0 <- types
	m, err := mocks.Generate(mockFinder)
	expect(err).To.Be.Nil()
	expect(m).To.Have.Len(2)
	expect(m[0]).To.Equal(mockFor(expect, types[0]))
	expect(m[1]).To.Equal(mockFor(expect, types[1]))
}

func TestOutput(t *testing.T) {
	expect := expect.New(t)

	types := []*ast.TypeSpec{
		typeSpec(expect, `
  type Foo interface {
   Bar() int
  }`),
		typeSpec(expect, `
  type Bar interface {
   Foo(foo string)
   Baz()
  }`),
	}

	mockFinder := newMockTypeFinder()
	mockFinder.ExportedTypesOutput.ret0 <- types
	m, err := mocks.Generate(mockFinder)
	expect(err).To.Be.Nil()

	buf := bytes.Buffer{}
	m.Output("foo", 100, &buf)

	// TODO: For some reason, functions are coming out without
	// whitespace between them.  We need to figure that out.
	expected, err := format.Source([]byte(`
 package foo
 
 type mockFoo struct {
  BarCalled chan bool
  BarOutput struct {
   ret0 chan int
  }
 }
 
 func newMockFoo() *mockFoo {
  m := &mockFoo{}
  m.BarCalled = make(chan bool, 100)
  m.BarOutput.ret0 = make(chan int, 100)
  return m
 } 
 func (m *mockFoo) Bar() int {
  m.BarCalled <- true
  return <-m.BarOutput.ret0
 }
 
 type mockBar struct {
  FooCalled chan bool
  FooInput struct {
   foo chan string
  }
  BazCalled chan bool
 }
 
 func newMockBar() *mockBar {
  m := &mockBar{}
  m.FooCalled = make(chan bool, 100)
  m.FooInput.foo = make(chan string, 100)
  m.BazCalled = make(chan bool, 100)
  return m
 }
 func (m *mockBar) Foo(foo string) {
  m.FooCalled <- true
  m.FooInput.foo <- foo
 } 
 func (m *mockBar) Baz() {
  m.BazCalled <- true
 }
 `))
	expect(err).To.Be.Nil()
	expect(buf.String()).To.Equal(string(expected))
}

func mockFor(expect func(interface{}) *expect.Expect, spec *ast.TypeSpec) mocks.Mock {
	m, err := mocks.For(spec)
	expect(err).To.Be.Nil()
	return m
}
