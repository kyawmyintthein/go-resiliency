package retrier

import (
	"errors"
	"testing"
)

type MyCustomError struct {
	message string
}

func (e *MyCustomError) Error() string {
	return e.message
}

var (
	errFoo    = errors.New("FOO")
	errBar    = errors.New("BAR")
	errBaz    = errors.New("BAZ")
	errCustom = &MyCustomError{message: "CustomError"}
)

func TestDefaultClassifier(t *testing.T) {
	c := DefaultClassifier{}

	if c.Classify(nil) != Succeed {
		t.Error("default misclassified nil")
	}

	if c.Classify(errFoo) != Retry {
		t.Error("default misclassified foo")
	}
	if c.Classify(errBar) != Retry {
		t.Error("default misclassified bar")
	}
	if c.Classify(errBaz) != Retry {
		t.Error("default misclassified baz")
	}
}

func TestWhitelistClassifier(t *testing.T) {
	c := NewWhiltelistClassifier([]error{errFoo, errBar, errCustom})

	if c.Classify(nil) != Succeed {
		t.Error("whitelist misclassified nil")
	}

	if c.Classify(errFoo) != Retry {
		t.Error("whitelist misclassified foo")
	}
	if c.Classify(errBar) != Retry {
		t.Error("whitelist misclassified bar")
	}

	if c.Classify(&MyCustomError{}) != Fail {
		t.Error("blacklist misclassified baz")
	}

	if c.Classify(errBaz) != Fail {
		t.Error("whitelist misclassified baz")
	}
}

func TestBlacklistClassifier(t *testing.T) {
	c := NewBlacklistClassifier([]error{errBar})

	if c.Classify(nil) != Succeed {
		t.Error("blacklist misclassified nil")
	}

	if c.Classify(errFoo) != Retry {
		t.Error("blacklist misclassified foo")
	}
	if c.Classify(errBar) != Fail {
		t.Error("blacklist misclassified bar")
	}
	if c.Classify(errBaz) != Retry {
		t.Error("blacklist misclassified baz")
	}
	if c.Classify(errCustom) != Retry {
		t.Error("blacklist misclassified baz")
	}
}
