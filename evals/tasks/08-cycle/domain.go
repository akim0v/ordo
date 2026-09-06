package main

type A struct{ b *B }

type B struct{ a *A }

// NewA depends on a *B.
func NewA(b *B) *A { return &A{b: b} }

// NewBWithA depends back on an *A, which closes a cycle.
func NewBWithA(a *A) *B { return &B{a: a} }

// NewB is the standalone constructor for a *B, with no dependency on *A.
func NewB() *B { return &B{} }

func main() {}
