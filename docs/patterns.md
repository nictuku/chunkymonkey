Patterns
========

This document is intended to convey the fundamental design patterns employed
within chunkymonkey.

Some names can probably be improved - please suggest better ones.


Embedded Encapsulated Value
---------------------------

This pattern is intended to allow one type to be directly embedded within
another, without the need to allocate a separate region of memory and address
it via a pointer.

Benefits of this include:

*   Clear ownership of the embedded value.
*   One less malloc() and free() per instance of the containing type.
*   (Potentially?) Better CPU L1/L2 cache performance - when one of embedded or
    containing value is loaded into a CPU cache, then the other will likely be
    loaded as well.

Disadvantages of this technique against using a pointer to contain the embedded
value include:

*   The contained value cannot have shared ownership.
*   The `nil` pointer value cannot be used to indicate some sort of absence of
    the embedded value.

Example: a type `Bar` in package `b` is to be embedded in type `Foo` from
package `f`.  `Bar` contains unexported fields whose 'default' value is not
regarded as 'initialized', and so `Bar` cannot be fully initialized outside of
package b.

    package b

    type Bar struct {
        privateValue int
    }

    func (bar *Bar) Init(barValue int) {
        bar.privateValue = barValue
    }

By virtue of `Bar.Init()`, The Foo type can then embed `Bar` within itself,
*and* initialize it correctly:

    package f

    type Foo struct {
        bar Bar
        foo string
    }

    func NewFoo(barValue int, fooValue string) (foo *Foo) {
        foo = new(Foo)
        foo.bar.Init(barValue)
        foo.foo = fooValue
        return
    }

Note that if `Foo` itself was intended to be embeddable like `Bar`, then it too
would need to provide a `Foo.Init()` method. To save coding, `NewFoo()` can be
defined in terms of `Foo.Init()`.

Note that this pattern also works for arrays or slices of the embeddable type.
