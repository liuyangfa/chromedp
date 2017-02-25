package chromedp

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/knq/chromedp/cdp"
	"github.com/knq/chromedp/cdp/input"
)

func TestMouseClickXY(t *testing.T) {
	t.Parallel()

	c := testAllocate(t, "input.html")
	defer c.Release()

	var err error
	err = c.Run(defaultContext, Sleep(time.Millisecond*100))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		x, y int64
	}{
		{100, 100},
		{0, 0},
		{9999, 100},
		{100, 9999},
	}

	for i, test := range tests {
		err = c.Run(defaultContext, MouseClickXY(test.x, test.y))
		if err != nil {
			t.Fatalf("test %d got error: %v", i, err)
		}

		time.Sleep(time.Millisecond * 100)

		var xstr, ystr string
		err = c.Run(defaultContext, Value("#input1", &xstr, ByID))
		if err != nil {
			t.Fatalf("test %d got error: %v", i, err)
		}
		x, err := strconv.ParseInt(xstr, 10, 64)
		if err != nil {
			t.Fatalf("test %d got error: %v", i, err)
		}
		if x != test.x {
			t.Fatalf("test %d expected x to be: %d, got: %d", i, test.x, x)
		}

		err = c.Run(defaultContext, Value("#input2", &ystr, ByID))
		if err != nil {
			t.Fatalf("test %d got error: %v", i, err)
		}
		y, err := strconv.ParseInt(ystr, 10, 64)
		if err != nil {
			t.Fatalf("test %d got error: %v", i, err)
		}
		if y != test.y {
			t.Fatalf("test %d expected y to be: %d, got: %d", i, test.y, y)
		}
	}
}

func TestMouseClickNode(t *testing.T) {
	tests := []struct {
		sel, exp string
		opt      MouseOption
		by       QueryOption
	}{
		{"button2", "foo", ButtonType(input.ButtonNone), ByID},
		{"button2", "bar", ButtonType(input.ButtonLeft), ByID},
		{"button2", "bar-middle", ButtonType(input.ButtonMiddle), ByID},
		{"input3", "bar-right", ButtonType(input.ButtonRight), ByID},
		{"input3", "bar-right", ButtonModifiers(input.ModifierNone), ByID},
		{"input3", "bar-right", Button("right"), ByID},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			t.Parallel()

			c := testAllocate(t, "input.html")
			defer c.Release()

			var err error
			var nodes []*cdp.Node
			err = c.Run(defaultContext, Nodes(test.sel, &nodes, test.by))
			if err != nil {
				t.Fatalf("got error: %v", err)
			}
			if len(nodes) != 1 {
				t.Fatalf("expected nodes to have exactly 1 element, got: %d", len(nodes))
			}

			err = c.Run(defaultContext, MouseClickNode(nodes[0], test.opt))
			if err != nil {
				t.Fatalf("got error: %v", err)
			}

			time.Sleep(time.Millisecond * 50)

			var value string
			err = c.Run(defaultContext, Value("#input3", &value, ByID))
			if err != nil {
				t.Fatalf("got error: %v", err)
			}
			if value != test.exp {
				t.Fatalf("expected to have value %s, got: %s", test.exp, value)
			}
		})
	}
}

func TestMouseClickOffscreenNode(t *testing.T) {
	tests := []struct {
		sel string
		exp int
		by  QueryOption
	}{
		{"#button3", 0, ByID},
		{"#button3", 2, ByID},
		{"#button3", 10, ByID},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			t.Parallel()

			c := testAllocate(t, "input.html")
			defer c.Release()

			var err error
			var nodes []*cdp.Node
			err = c.Run(defaultContext, Nodes(test.sel, &nodes, test.by))
			if err != nil {
				t.Fatalf("got error: %v", err)
			}
			if len(nodes) != 1 {
				t.Fatalf("expected nodes to have exactly 1 element, got: %d", len(nodes))
			}

			var ok bool
			err = c.Run(defaultContext, EvaluateAsDevTools(fmt.Sprintf(isOnViewJS, nodes[0].FullXPath()), &ok))
			if err != nil {
				t.Fatalf("got error: %v", err)
			}
			if ok {
				t.Fatal("expected node to be offscreen")
			}

			for i := test.exp; i > 0; i-- {
				err = c.Run(defaultContext, MouseClickNode(nodes[0]))
				if err != nil {
					t.Fatalf("got error: %v", err)
				}
			}

			time.Sleep(time.Millisecond * 100)

			var value int
			err = c.Run(defaultContext, Evaluate("window.document.test_i", &value))
			if err != nil {
				t.Fatalf("got error: %v", err)
			}
			if value != test.exp {
				t.Fatalf("expected to have value %d, got: %d", test.exp, value)
			}
		})
	}
}

func TestKeyAction(t *testing.T) {
	tests := []struct {
		sel, exp string
		by       QueryOption
	}{
		{"#input4", "foo", ByID},
		{"#input4", "foo and bar", ByID},
		{"#input4", "1234567890", ByID},
		{"#input4", "~!@#$%^&*()_+=[];'", ByID},
		{"#input4", "你", ByID},
		{"#input4", "\n\nfoo\n\nbar\n\n", ByID},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			t.Parallel()

			c := testAllocate(t, "input.html")
			defer c.Release()

			var err error
			var nodes []*cdp.Node
			err = c.Run(defaultContext, Nodes(test.sel, &nodes, test.by))
			if err != nil {
				t.Fatalf("got error: %v", err)
			}
			if len(nodes) != 1 {
				t.Fatalf("expected nodes to have exactly 1 element, got: %d", len(nodes))
			}

			err = c.Run(defaultContext, Focus(test.sel, test.by))
			if err != nil {
				t.Fatalf("got error: %v", err)
			}

			err = c.Run(defaultContext, KeyAction(test.exp))
			if err != nil {
				t.Fatalf("got error: %v", err)
			}

			var value string
			err = c.Run(defaultContext, Value(test.sel, &value, test.by))
			if err != nil {
				t.Fatalf("got error: %v", err)
			}
			if value != test.exp {
				t.Fatalf("expected to have value %s, got: %s", test.exp, value)
			}
		})
	}
}

func TestKeyActionNode(t *testing.T) {
	tests := []struct {
		sel, exp string
		by       QueryOption
	}{
		{"#input4", "foo", ByID},
		{"#input4", "foo and bar", ByID},
		{"#input4", "1234567890", ByID},
		{"#input4", "~!@#$%^&*()_+=[];'", ByID},
		{"#input4", "你", ByID},
		{"#input4", "\n\nfoo\n\nbar\n\n", ByID},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			t.Parallel()

			c := testAllocate(t, "input.html")
			defer c.Release()

			var err error
			var nodes []*cdp.Node
			err = c.Run(defaultContext, Nodes(test.sel, &nodes, test.by))
			if err != nil {
				t.Fatalf("got error: %v", err)
			}
			if len(nodes) != 1 {
				t.Fatalf("expected nodes to have exactly 1 element, got: %d", len(nodes))
			}

			err = c.Run(defaultContext, KeyActionNode(nodes[0], test.exp))
			if err != nil {
				t.Fatalf("got error: %v", err)
			}

			var value string
			err = c.Run(defaultContext, Value(test.sel, &value, test.by))
			if err != nil {
				t.Fatalf("got error: %v", err)
			}
			if value != test.exp {
				t.Fatalf("expected to have value %s, got: %s", test.exp, value)
			}
		})
	}
}