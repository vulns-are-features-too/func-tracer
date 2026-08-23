package fakes

import (
	"context"

	"github.com/vulns-are-features-too/func-tracer/src/model"
)

type funcCall struct {
	caller *model.Symbol
	callee *model.Symbol
}

type fakeSession struct {
	calls []funcCall
}

// Session fake for testing.
//
//nolint:revive // unexported-return
func Session() *fakeSession {
	return &fakeSession{calls: make([]funcCall, 0)}
}

// AddCall to calls list.
func (s *fakeSession) AddCall(caller, callee *model.Symbol) {
	s.calls = append(s.calls, funcCall{caller, callee})
}

// HasCaller in calls list.
func (s *fakeSession) HasCaller(callee *model.Symbol) bool {
	for _, c := range s.calls {
		if callee.ID == c.callee.ID {
			return true
		}
	}

	return false
}

// HasCallee in calls list.
func (s *fakeSession) HasCallee(caller *model.Symbol) bool {
	for _, c := range s.calls {
		if caller.ID == c.caller.ID {
			return true
		}
	}

	return false
}

// References of function at location.
func (s *fakeSession) References(
	_ context.Context,
	location model.Location,
) ([]model.Location, error) {
	results := make([]model.Location, 0)

	for _, call := range s.calls {
		if call.callee.Location.Range == location.Range {
			results = append(results, call.caller.Location)
		}
	}

	return results, nil
}

// Close does nothing.
func (*fakeSession) Close(context.Context) {}
