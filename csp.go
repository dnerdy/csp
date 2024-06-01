package csp

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// CSP Paper: https://www.cs.cmu.edu/~crary/819-f09/Hoare78.pdf

// 2.3 Input and Output commands
//
// (Don't worry! We discuss sections 2.1 and 2.2 below.)
//
// Input and output are the "?" and "!" operators.

// cardreader?cardimage is a channel read

func simpleRead() {
	cardimage = <-cardreader
}

// lineprinter!lineimage is a channel write

func simpleWrite() {
	lineprinter <- lineimage
}

// Input and output sources and destinations in CSP function exactly like
// unbuffered channels in Go--the type of channel created in Go when writing
// `make(chan any)`. A channel read occurs if a channel writer is blocked
// trying to perform a write. If there is no blocker writer, the read blocks.
// When a writer attempts to write, if there's a blocked reader the write
// occurs. Else, the writer blocks (waiting for a reader to come by and read
// the value being written).

// Section 2.1 is all about the "||" operator, the "parallel" operator. It
// works like an error group:
//
// [cardreader?cardimage||lineprinter!lineimage]

func parallel() error {
	var g *errgroup.Group
	var ctx context.Context

	g, ctx = errgroup.WithContext(ctx)
	g.Go(func() error {
		cardimage = <-cardreader
		return nil
	})
	g.Go(func() error {
		lineprinter <- lineimage
		return nil
	})
	return g.Wait()
}

// In CSP, the goroutines in a parallel command may NOT mutate shared state.
// That is, a variable that appears in the left-hand-side (LHS) of an
// assignment in one goroutine may not also appear in the LHS of an
// assignment in another goroutine.

// 2.2 Assignment

func assignmentExamples() {
	// Example: x := x + 1 -- same in Go:
	x = x + 1

	// Example: x, y = y, x -- same in Go:
	x, y = y, x

	// All the rest of the examples show pattern matching, which Go
	// doesn't have but that can be approximated with structs.

	// `x := cons(left, right)` is like:
	type cons struct {
		left  int
		right int
	}

	x_ := cons{left, right}

	// `cons(left, right) := x` is like:

	var result cons = x_

	left = result.left
	right = result.right

	// All the rest of the examples are the same.

	// Note that the last example:
	//   insert(n) := has(n)
	// is prevented by the compiler in Go, since Go is statically typed.
	//
	// In spirit `insert(n) := has(n)` is similar to:
	type has struct{ n int }
	type insert struct{ n int }

	var rhs any = has{1} // Note: "rhs" is "right-hand-side"

	lhs = rhs.(insert) // This will panic
}

// 2.3 Input and Output commands
//
// These are all the same as the channel reads and writes shown in simpleRead
// and simpleWrite above.

// Of note, console(i) is array indexing:

func arrayIndex() {
	c = console[i]
}

// And the declaration `buffer:(1..80)character` is:

func arrayDeclaration() {
	buffer := make([]string, 80)
}

// 2.4 Alternative and Repetitive commands
//
// This is where things get spicy!
//
// Repetitive commands are like `for`.
// Alternative commands are like `select`.
//
// Alternative cases can have guards. Go doesn't have this concept, but we'll
// show how the concept would function using a succession of Go snippets.

// First, let's tackle alternative commands: code surrounded by "[" and "]"
// that uses the "▯" character to separate each case.
//
// [c:character; west?c -> east!c ▯ c2:character; north?c2 -> south!c2] corresponds to:

func simpleAlternative() {
	select {
	case c := <-west:
		east <- c
	case c2 := <-north:
		south <- c2
	}
}

// `<-west` and `<-north` are the "input" parts of the guards. Each guard can
// have one (optional) input.

// Note that in CSP, just like in Go, if both `west` and `north` are ready,
// one of the cases is chosen at random.
//
// ANOTHER IMPORTANT DIFFERENCE: If a channel is closed, CSP excludes the
// corresponding case(s) from the select.

// Cool. Now let's look at a guard example that has boolean criteria in
// addition to channel reads.
//
// [c:character; x > y; west?c -> east!c ▯ c2:character; y > x; north?c2 -> south!c2]
//
// This would be a select with special semantics like:

func alternativeWithBooleanCriteria() {
	select {
	case c := <-west: // when x > y
		east <- c
	case c2 := <-north: // when y > x
		south <- c2
	}
}

// The way this functions is: when entering the select, only keep the cases
// where the "when" conditions are true; the other cases are removed.
//
// When x > y, we get:

func alternativeWithBooleanCriteriaXGreaterThanY() {
	select {
	case c := <-west:
		east <- c
	}
}

// What if there isn't any matching criteria, e.g. when x == y?

func alternativeXGreaterThanY() {
	select {}
}

// In Go, this blocks forever. In CSP, it does nothing, or
// IN THE CASE OF REPETITION, it breaks out of the outer loop.

// So, let's talk about repetition, the "*" operator.
//
// *[c:character; x > y; west?c -> east!c ▯ c2:character; y > x; north?c2 -> south!c2]
//
// This is the same alternate command as before but with a "*" before the opening "[".
//
// This corresponds to:

func repetition() {
LOOP:
	for {
		select {
		case c := <-west: // when x > y
			east <- c
		case c2 := <-north: // when y > x
			south <- c2
		default: // when no cases left
			break LOOP
		}
	}
}

// When x > y, we get:

func repetitionXGreaterThanY() {
	for {
		select {
		case c := <-west:
			east <- c
		}
	}
}

// And when x == y, we get:

func alternativeWithXEqualY() {
LOOP:
	for {
		select {
		default:
			break LOOP
		}
	}
}

// How about when x > y and `west` is closed?
//
// Recall that in CSP, a case is REMOVED from the select in if the channel
// being read from is closed. So again we get:

func alternativeWithWestClosed() {
LOOP:
	for {
		select {
		default:
			break LOOP
		}
	}
}

// Above I mentioned that the input part of a guard is optional. What does a
// guard without input look like?
//
// [x > y -> m := x ▯ y > x -> m := y]
//
// It would be:

func nonInputGuards() {
	select {
	case <-alwaysReady: // when x > y
		m = x
	case <-alwaysReady: // when y > x
		m = y
	}
}

// Finally, note that guards can contain multiple boolean expressions, e.g.
// [x > y; x + y < 10 -> m := x]
//
// The boolean expressions are evaluated left to right. If any evaluates to
// false, the guard is false (i.e. the statements separated by ";" use
// short-circuited && semantics).

// Misc. CSP array-oriented features.
//
// *[(i:1..3)continue(i); console(i)?c -> X!(i, c); console(i)!ack(); continue() := (c != SIGN_OFF)]
//
// Can be written as:

func arrayExampleOne() {
	continue_ := make([]bool, 3)

	for {
		select {
		case c := <-console[0]: // when continue[0] == true
			X <- xinput{i, c}
			console[i] <- ack{}
			continue_[i] = c != SIGN_OFF
		case c := <-console[1]: // when continue[1] == true
			X <- xinput{i, c}
			console[i] <- ack{}
			continue_[i] = c != SIGN_OFF
		case c := <-console[2]: // when continue[2] == true
			X <- xinput{i, c}
			console[i] <- ack{}
			continue_[i] = c != SIGN_OFF
		}
	}
}

// This would be like having a `for select` operator in Go like:

func fictionalForSelectInGo() {
	continue_ := make([]bool, 3)

LOOP:
	for {
		/* for */ select /* i := range 3 */ {
		case c := <-console[i]: // when continue[i] == true
			X <- xinput{i, c}
			console[i] <- ack{}
			continue_[i] = c != SIGN_OFF
		default: // when no cases match
			break LOOP
		}
	}
}

//////////////////////////////////////////////
// Declarations so the examples above compile.

var cardreader = make(chan string)
var cardimage string

var lineprinter = make(chan string)
var lineimage string

var x int
var y int
var m int
var c any
var lhs any

var left int
var right int

var i int
var console []chan any
var west chan any
var east chan any
var north chan any
var south chan any
var alwaysReady chan any
var X chan struct {
	i int
	c any
}

type ack struct{}
type xinput struct {
	i int
	c any
}

const SIGN_OFF = "q"
