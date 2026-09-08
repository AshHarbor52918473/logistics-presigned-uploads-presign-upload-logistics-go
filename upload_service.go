package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

const maxAssetBytes int64 = 10 << 20

type shipmentEvent string

const (
	eventDelivered shipmentEvent = "delivered"
	eventException shipmentEvent = "exception"
)

type uploadRequest struct {
	ShipmentID  string        `json:"shipment_id"`
	Event       shipmentEvent `json:"event"`
	AssetKind   string        `json:"asset_kind"`
	Filename    string        `json:"filename"`
	ContentType string        `json:"content_type"`
	SizeBytes   int64         `json:"size_bytes"`
}

type uploadGrant struct {
	UploadURL string `json:"upload_url"`
	Method    string `json:"method"`
	ObjectKey string `json:"object_key"`
	ExpiresIn int    `json:"expires_seconds"`
}

type putSigner interface {
	presignPut(context.Context, string, string, presignRequest) (presignData, error)
}

type uploadService struct {
	bucket string
	signer putSigner
}

func (s uploadService) grant(ctx context.Context, in uploadRequest) (uploadGrant, error) {
	if err := validateUpload(in); err != nil {
		return uploadGrant{}, err
	}
	key := objectKey(in)
	expires := 600
	signed, err := s.signer.presignPut(ctx, s.bucket, key, presignRequest{
		Op:             "put",
		ExpiresSeconds: expires,
		ContentType:    in.ContentType,
		MaxBytes:       in.SizeBytes,
		IdempotencyKey: requestID(in),
	})
	if err != nil {
		return uploadGrant{}, err
	}
	return uploadGrant{UploadURL: signed.URL, Method: "PUT", ObjectKey: key, ExpiresIn: expires}, nil
}

func validateUpload(in uploadRequest) error {
	if strings.TrimSpace(in.ShipmentID) == "" || strings.ContainsAny(in.ShipmentID, "/\\") {
		return errors.New("shipment_id must be a non-empty path-safe value")
	}
	allowed := (in.Event == eventDelivered && in.AssetKind == "proof_of_delivery") ||
		(in.Event == eventException && in.AssetKind == "exception_evidence")
	if !allowed {
		return errors.New("asset_kind does not match the shipment event")
	}
	if in.ContentType != "image/jpeg" && in.ContentType != "image/png" && in.ContentType != "application/pdf" {
		return errors.New("content_type must be image/jpeg, image/png, or application/pdf")
	}
	if in.SizeBytes <= 0 || in.SizeBytes > maxAssetBytes {
		return fmt.Errorf("size_bytes must be between 1 and %d", maxAssetBytes)
	}
	name := filepath.Base(strings.TrimSpace(in.Filename))
	if name == "." || name == "" {
		return errors.New("filename is required")
	}
	return nil
}

func objectKey(in uploadRequest) string {
	return fmt.Sprintf("shipments/%s/%s/%s", in.ShipmentID, in.AssetKind, filepath.Base(in.Filename))
}

func requestID(in uploadRequest) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		in.ShipmentID, string(in.Event), in.AssetKind, in.Filename, in.ContentType, fmt.Sprint(in.SizeBytes),
	}, "\x00")))
	return "shipment-asset-" + hex.EncodeToString(sum[:16])
}
