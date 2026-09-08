package main

import (
	"context"
	"testing"
)

type recordingSigner struct {
	calls int
	key   string
	body  presignRequest
}

func (s *recordingSigner) presignPut(_ context.Context, _ string, key string, body presignRequest) (presignData, error) {
	s.calls++
	s.key = key
	s.body = body
	return presignData{URL: "https://uploads.example/signed"}, nil
}

func TestUploadGrantFollowsShipmentEventPolicy(t *testing.T) {
	tests := []struct {
		name      string
		event     shipmentEvent
		assetKind string
		wantOK    bool
	}{
		{name: "delivery accepts POD", event: eventDelivered, assetKind: "proof_of_delivery", wantOK: true},
		{name: "exception accepts evidence", event: eventException, assetKind: "exception_evidence", wantOK: true},
		{name: "delivery rejects exception evidence", event: eventDelivered, assetKind: "exception_evidence"},
		{name: "exception rejects POD", event: eventException, assetKind: "proof_of_delivery"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			signer := &recordingSigner{}
			service := uploadService{bucket: "logistics-assets", signer: signer}
			grant, err := service.grant(context.Background(), uploadRequest{
				ShipmentID: "SHP-1042",
				Event:      tc.event, AssetKind: tc.assetKind,
				Filename: "dock-photo.jpg", ContentType: "image/jpeg", SizeBytes: 2048,
			})
			if tc.wantOK && err != nil {
				t.Fatalf("grant() error = %v", err)
			}
			if !tc.wantOK {
				if err == nil {
					t.Fatal("grant() expected policy rejection")
				}
				if signer.calls != 0 {
					t.Fatalf("signer called %d times for rejected input", signer.calls)
				}
				return
			}
			if grant.Method != "PUT" || signer.body.Op != "put" {
				t.Fatalf("grant method = %q, signer op = %q", grant.Method, signer.body.Op)
			}
			if signer.body.IdempotencyKey == "" || signer.key == "" {
				t.Fatal("accepted upload must have an object key and idempotency key")
			}
		})
	}
}
