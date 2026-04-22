package handler_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/payment/handler"
)

type fakeCallbackService struct {
	markPaymentSuccessCalled int
	lastPaymentOrderNo       string
	lastRawPayload           []byte
	err                      error
}

func (f *fakeCallbackService) MarkPaymentSuccess(ctx context.Context, paymentOrderNo string, rawPayload []byte) error {
	f.markPaymentSuccessCalled++
	f.lastPaymentOrderNo = paymentOrderNo
	f.lastRawPayload = rawPayload
	return f.err
}

type fakeVerifier struct {
	paymentOrderNo string
	err            error
}

func (f *fakeVerifier) VerifyCallback(r *http.Request) (string, error) {
	return f.paymentOrderNo, f.err
}

func TestCallbackHandlerReturnsOKOnSuccess(t *testing.T) {
	svc := &fakeCallbackService{}
	verifier := &fakeVerifier{paymentOrderNo: "P12345"}
	h := handler.NewCallbackHandler(svc, verifier)

	body := []byte(`{"status":"success"}`)
	req := httptest.NewRequest(http.MethodPost, "/payments/callback", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if svc.markPaymentSuccessCalled != 1 {
		t.Fatalf("expected MarkPaymentSuccess called once, got %d", svc.markPaymentSuccessCalled)
	}
	if svc.lastPaymentOrderNo != "P12345" {
		t.Fatalf("expected payment order no P12345, got %s", svc.lastPaymentOrderNo)
	}
	if string(svc.lastRawPayload) != `{"status":"success"}` {
		t.Fatalf("expected raw payload, got %s", string(svc.lastRawPayload))
	}
}

func TestCallbackHandlerReturnsErrorWhenVerifierFails(t *testing.T) {
	svc := &fakeCallbackService{}
	verifier := &fakeVerifier{err: errorsx.ErrBadRequest}
	h := handler.NewCallbackHandler(svc, verifier)

	req := httptest.NewRequest(http.MethodPost, "/payments/callback", nil)
	w := httptest.NewRecorder()

	h.Handle(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if svc.markPaymentSuccessCalled != 0 {
		t.Fatalf("expected MarkPaymentSuccess not called, got %d", svc.markPaymentSuccessCalled)
	}
}

func TestCallbackHandlerReturnsErrorWhenServiceFails(t *testing.T) {
	svc := &fakeCallbackService{err: errorsx.ErrInternal}
	verifier := &fakeVerifier{paymentOrderNo: "P12345"}
	h := handler.NewCallbackHandler(svc, verifier)

	req := httptest.NewRequest(http.MethodPost, "/payments/callback", bytes.NewReader([]byte(`{}`)))
	w := httptest.NewRecorder()

	h.Handle(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestCallbackHandlerIdempotentDuplicateStillReturnsOK(t *testing.T) {
	svc := &fakeCallbackService{}
	verifier := &fakeVerifier{paymentOrderNo: "P12345"}
	h := handler.NewCallbackHandler(svc, verifier)

	req := httptest.NewRequest(http.MethodPost, "/payments/callback", bytes.NewReader([]byte(`{}`)))
	w := httptest.NewRecorder()
	h.Handle(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("first call: expected 200, got %d", w.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/payments/callback", bytes.NewReader([]byte(`{}`)))
	w2 := httptest.NewRecorder()
	h.Handle(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("duplicate call: expected 200, got %d", w2.Code)
	}
	if svc.markPaymentSuccessCalled != 2 {
		t.Fatalf("expected MarkPaymentSuccess called twice (idempotency is in service layer), got %d", svc.markPaymentSuccessCalled)
	}
}
