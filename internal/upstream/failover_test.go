package upstream

import (
	"context"
	"errors"
	"testing"
)

type fakeAPI struct {
	loginErr error
	usageErr error
	logins   int
}

func (f *fakeAPI) Login(context.Context) error   { f.logins++; return f.loginErr }
func (f *fakeAPI) Refresh(context.Context) error { return nil }
func (f *fakeAPI) ListUsage(context.Context, UsageQuery) (Page[UsageRecord], error) {
	return Page[UsageRecord]{}, f.usageErr
}
func (f *fakeAPI) ListLedger(context.Context, LedgerQuery) (Page[LedgerEntry], error) {
	return Page[LedgerEntry]{}, nil
}

func TestFailoverMovesToNextHostOnServerError(t *testing.T) {
	first := &fakeAPI{loginErr: &Error{Kind: ErrorServer, Status: 503, Err: errors.New("down")}}
	second := &fakeAPI{}
	f := NewFailover([]Endpoint{{URL: "primary", API: first}, {URL: "backup", API: second}})
	if err := f.Login(context.Background()); err != nil {
		t.Fatal(err)
	}
	if f.CurrentHost() != "backup" || first.logins != 1 || second.logins != 1 {
		t.Fatalf("host=%q logins=%d/%d", f.CurrentHost(), first.logins, second.logins)
	}
}

func TestFailoverStopsOnAuthenticationError(t *testing.T) {
	first := &fakeAPI{loginErr: &Error{Kind: ErrorAuth, Status: 401, Err: errors.New("bad credentials")}}
	second := &fakeAPI{}
	f := NewFailover([]Endpoint{{URL: "primary", API: first}, {URL: "backup", API: second}})
	if err := f.Login(context.Background()); !IsKind(err, ErrorAuth) {
		t.Fatalf("error=%v", err)
	}
	if second.logins != 0 {
		t.Fatal("authentication error triggered blind host switching")
	}
}

func TestFailoverRequestSwitchesAfterTransportFailure(t *testing.T) {
	first := &fakeAPI{usageErr: &Error{Kind: ErrorTransport, Err: errors.New("network")}}
	second := &fakeAPI{}
	f := NewFailover([]Endpoint{{URL: "primary", API: first}, {URL: "backup", API: second}})
	if err := f.Login(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := f.ListUsage(context.Background(), UsageQuery{}); err != nil {
		t.Fatal(err)
	}
	if f.CurrentHost() != "backup" {
		t.Fatalf("host=%q", f.CurrentHost())
	}
}
