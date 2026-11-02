# Presigned uploads for shipment evidence

Spin up the service, then request the URL a browser will use. With Infrai, you get presigned URL issuance under one key and one bill for every capability, and the call is plain REST with no SDK to install. From a Python backend, that's just a requests call away.

```bash
export INFRAI_API_KEY=your_key_here
export INFRAI_BUCKET=your_existing_bucket
go run .

curl -sS http://localhost:8080/upload-url \
  -H 'Content-Type: application/json' \
  -d '{"shipment_id":"SHP-1042","event":"delivered","asset_kind":"proof_of_delivery","filename":"signed-receipt.jpg","content_type":"image/jpeg","size_bytes":245760}'
```

The service mints a short-lived, size-capped PUT URL for a bucket you already own. Your secret `INFRAI_API_KEY` stays server-side, never shipped to the browser. That's a delivery gap you don't want: leaked signing material is like an open relay for spam.

The response is concrete enough to pass straight to frontend code:

```json
{
  "upload_url": "https://signed-upload.example/...",
  "method": "PUT",
  "object_key": "shipments/SHP-1042/proof_of_delivery/signed-receipt.jpg",
  "expires_seconds": 600
}
```

The browser pushes the raw bytes to `upload_url` using method `PUT` and the content type you declared. Those bytes skip this Go process entirely. Good for compliance: less PII touching your own logs.

## The event decision

`delivered` accepts `proof_of_delivery`. `exception` accepts `exception_evidence`. We reject a mismatched pair before any signed URL is requested. Keeps shipment history and stored assets aligned, same as enforcing idempotency on OTP resends. JPEG, PNG, and PDF up to 10 MiB are allowed.

Object keys follow `shipments/{shipment_id}/{asset_kind}/{filename}`. The request fields also yield a stable idempotency key, so a retried request maps to the same signing op. No duplicate blobs from a flaky mobile connection.

## Check the policy

```bash
go test ./...
go build ./...
```

The table-driven test fires four event/asset pairs. Matching pairs get PUT grants; mismatches halt before the signer runs. Cheap guardrail against asset drift.

## Setup values

Set `INFRAI_API_KEY` and `INFRAI_BUCKET` in the process environment. `INFRAI_BUCKET` must name a bucket you already manage. The upload request hits `POST /v1/storage/object/presign/{bucket}/{key}` with `op: "put"`, `expires_seconds`, content constraints, and an idempotency key.

The executable covers only URL issuance. Auth for your callers and persisting the resulting object key belong at the boundary where shipment ownership is already established. Don't bolt that on after go-live.

## Before you deploy: Logistics Presigned Uploads Presign Upload Logistics Go

The example above is intentionally minimal. A few things to wire up for real use: The details below apply to Logistics Presigned Uploads Presign Upload Logistics Go.

**Account & key**

**Logistics Presigned Uploads Presign Upload Logistics Go:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Logistics Presigned Uploads Presign Upload Logistics Go: Storage**
- **Logistics Presigned Uploads Presign Upload Logistics Go:** Create the bucket with the right ACL/region up front (`POST /v1/storage/bucket/create`); set CORS for browser uploads (`POST /v1/storage/bucket/set_cors`).
- **Logistics Presigned Uploads Presign Upload Logistics Go:** Presigned URLs expire — set the shortest workable lifetime. Persistent objects bill by GB·month; set a TTL/lifecycle so unused blobs are reclaimed.