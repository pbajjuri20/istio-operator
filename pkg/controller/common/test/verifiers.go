package test

import (
	"fmt"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"
)

// Verify creates a new ActionVerifierFactory for the given verb.
// Use "*" to match any verb.  Unless the filter is narrowed, e.g. using On(),
// the ActionVerifier created by this factory will match all resource types in
// all namespaces (i.e. the other filter fields are initialized to "*").
func Verify(verb string) *ActionVerifierFactory {
	return &ActionVerifierFactory{
		AbstractActionFilter: AbstractActionFilter{
			Verb:        verb,
			Namespace:   "*",
			Name:        "*",
			Resource:    "*",
			Subresource: "*",
		},
	}
}

// ActionVerifierFactory is a factory for creating common verifiers
type ActionVerifierFactory struct {
	AbstractActionFilter
}

// On initializes the resource and subresource name to which the created
// verifier should apply.  resource parameter should be specified using a slash
// between resource an subresource, e.g. deployments/status.  Use "*" to match
// all resources.
func (f *ActionVerifierFactory) On(resource string) *ActionVerifierFactory {
    f.AbstractActionFilter.On(resource)
    return f
}

// In initializes the namespace whithin which the created verifier should apply.
// Use "*" to match all namespaces.
func (f *ActionVerifierFactory) In(namespace string) *ActionVerifierFactory {
    f.AbstractActionFilter.In(namespace)
    return f
}

// Named initializes the name of the resource to which the created verifier
// should apply.  Use "*" to match all names.
func (f *ActionVerifierFactory) Named(name string) *ActionVerifierFactory {
    f.AbstractActionFilter.Named(name)
    return f
}

// IsSeen returns an ActionVerifier that verifies the specified action has occurred.
func (f *ActionVerifierFactory) IsSeen() ActionVerifier {
	return NewSimpleActionVerifier(f.Verb, f.Resource, f.Subresource, f.Namespace, f.Name,
		func(action clienttesting.Action) (bool, error) {
			return true, nil
		})
}

// SeenCountIs returns an ActionVerifier that verifies the specified action has
// occurred exactly the expected number of times.
func (f *ActionVerifierFactory) SeenCountIs(expected int) ActionVerifier {
	return NewSimpleActionVerifier(f.Verb, f.Resource, f.Subresource, f.Namespace, f.Name,
		func(action clienttesting.Action) (bool, error) {
			expected--
			return expected == 0, nil
		})
}

// IsNotSeen returns an ActionVerifier that verifies the specified action has occurred.
// This should be the last verifier in a list of verifiers, as it will wait for the
// full timeout before returning success.
func (f *ActionVerifierFactory) IsNotSeen() ActionVerifier {
	verifier := &notSeenActionVerifier{SimpleActionVerifier: *NewSimpleActionVerifier(f.Verb, f.Resource, f.Subresource, f.Namespace, f.Name, nil)}
	verifier.Verify = func(action clienttesting.Action) (bool, error) {
		return true, fmt.Errorf("unexpected %s action occurred: %s", verifier.AbstractActionFilter.String(), action)
	}
	return verifier
}

// With creturns an ActionVerifier that verifies the filtered action using the
// specified verifier function.
func (f *ActionVerifierFactory) With(verifier ActionVerifierFunc) ActionVerifier {
	return NewSimpleActionVerifier(f.Verb, f.Resource, f.Subresource, f.Namespace, f.Name, verifier)
}

type notSeenActionVerifier struct {
	SimpleActionVerifier
}

func (v *notSeenActionVerifier) Wait(timeout time.Duration) bool {
	select {
	case <-v.Notify:
	case <-time.After(timeout):
		// no error on a timeout
	}
	if !v.HasFired() {
		v.fired = true
		close(v.Notify)
	}
	return false
}

/*
SimpleActionVerifier is a simple ActionVerifier that applies the validation
logic when verb/resource/subresource/name/namespace match an action.  The
verification logic is only executed once.  This can be used as the base for a
custom verifier by overriding the Handles() method, e.g.

	type CustomActionVerifier struct {
		test.SimpleActionVerifier
	}

	func (v *CustomActionVerifier) Handles(action clienttesting.Action) bool {
		if v.SimpleActionVerifier.Handles(action) {
			// custom handling logic
			return true
		}
		return false
	}

	customVerifier := &CustomActionVerifier{SimpleActionVerifier: test.VerifyAction(...)}
*/
type SimpleActionVerifier struct {
	AbstractActionFilter
	Verify ActionVerifierFunc
	fired  bool
	Notify chan struct{}
	t      *testing.T
}

var _ ActionVerifier = (*SimpleActionVerifier)(nil)

// NewSimpleActionVerifier returns a new ActionVerifier that filtering the
// specified verb, resource, etc., using the specified verifier function.
func NewSimpleActionVerifier(verb, resource, subresource, namespace, name string, verifier ActionVerifierFunc) *SimpleActionVerifier {
	return &SimpleActionVerifier{
		AbstractActionFilter: AbstractActionFilter{
			Namespace:   namespace,
			Name:        name,
			Verb:        verb,
			Resource:    resource,
			Subresource: subresource,
		},
		Verify: verifier,
		Notify: make(chan struct{}),
	}
}

// Handles returns true if the action matches the settings for this verifier
// (verb, resource, subresource, namespace, and name) and the verifier has not
// already been applied.
func (v *SimpleActionVerifier) Handles(action clienttesting.Action) bool {
	if v.fired {
		return false
	}
	return v.AbstractActionFilter.Handles(action)
}

// React to the action.  This method always returns false for handled, as it only
// verifies the action.  It does not perform the action.  If verification fails,
// it will register a fatal error with the test runner.
func (v *SimpleActionVerifier) React(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
	v.t.Helper()
	if handled, err := v.Verify(action); handled || err != nil {
		v.fired = true
		defer close(v.Notify)
		if err != nil {
			v.t.Fatal(err)
		}
	}
	return false, nil, nil
}

// Wait until the verification has completed.  Returns true if it timed out waiting for verification.
func (v *SimpleActionVerifier) Wait(timeout time.Duration) (timedout bool) {
	v.t.Helper()
	select {
	case <-v.Notify:
	case <-time.After(timeout):
		v.t.Errorf("verify %s timed out", v.AbstractActionFilter.String())
		return true
	}
	return false
}

// InjectTestRunner initializes the test runner for the verifier.
func (v *SimpleActionVerifier) InjectTestRunner(t *testing.T) {
	v.t = t
}

// HasFired returns true if this verifier has fired
func (v *SimpleActionVerifier) HasFired() bool {
	return v.fired
}

// VerifyActions is a list of ActionVerifier objects which are applied in order.
type VerifyActions []ActionVerifier

var _ ActionVerifier = (*VerifyActions)(nil)

// Handles tests the head of the list to see if the action should be verified.
func (v *VerifyActions) Handles(action clienttesting.Action) bool {
	return len(*v) > 0 && (*v)[0].Handles(action)
}

// React verifies the action using the ActionVerifier at that head of the list.
func (v *VerifyActions) React(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
	(*v)[0].React(action)
	if (*v)[0].HasFired() {
		*v = (*v)[1:]
	}
	return false, nil, nil
}

// Wait for all ActionVerifier elements in this list to complete.
func (v *VerifyActions) Wait(timeout time.Duration) (timedout bool) {
	start := time.Now()
	verifiers := (*v)[:]
	for _, verifier := range verifiers {
		if timedout := verifier.Wait(timeout - time.Now().Sub(start)); timedout {
			return true
		}
	}
	return false
}

// InjectTestRunner injects the test runner into each ActionVerifier in the list.
func (v *VerifyActions) InjectTestRunner(t *testing.T) {
	for _, verifier := range *v {
		verifier.InjectTestRunner(t)
	}
}

// HasFired returns true if this verifier has fired
func (v *VerifyActions) HasFired() bool {
	return len(*v) == 0
}
