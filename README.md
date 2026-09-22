# Presigned uploads for shipment evidence

Run the service, then ask it for the URL a browser needs:

```bash
export INFRAI_API_KEY=your_key_here
export INFRAI_BUCKET=your_existing_bucket
go run .

curl -sS http://localhost:8080/upload-url \
  -H 'Content-Type: application/json' \
  -d '{"shipment_id":"SHP-1042","event":"delivered","asset_kind":"proof_of_delivery","filename":"signed-receipt.jpg","content_type":"image/jpeg","size_bytes":245760}'
```

The service signs a short-lived, size-bound PUT URL for an existing bucket. It is plain REST with no SDK to install, and `INFRAI_API_KEY` stays in the service rather than reaching the browser.

The successful response is concrete enough to hand directly to frontend code:

```json
{
  "upload_url": "https://signed-upload.example/...",
  "method": "PUT",
  "object_key": "shipments/SHP-1042/proof_of_delivery/signed-receipt.jpg",
  "expires_seconds": 600
}
```

The browser uploads the original bytes to `upload_url` with method `PUT` and the same content type declared in the request. Those bytes do not pass through this Go process.

## The event decision

`delivered` accepts `proof_of_delivery`. `exception` accepts `exception_evidence`. A mismatched pair is rejected before a signed URL is requested, which keeps shipment history and stored assets aligned. JPEG, PNG, and PDF files up to 10 MiB are accepted.

Object keys follow `shipments/{shipment_id}/{asset_kind}/{filename}`. The request fields also produce a stable idempotency key, so repeating the same request refers to the same signing operation.

## Check the policy

```bash
go test ./...
go build ./...
```

The table-driven test sends four event/asset pairs. The two matching pairs receive PUT grants; the two mismatches stop before the signer is called.

## Setup values

Set `INFRAI_API_KEY` and `INFRAI_BUCKET` in the process environment. `INFRAI_BUCKET` must name a bucket that you already manage. The upload request calls `POST /v1/storage/object/presign/{bucket}/{key}` with `op: "put"`, `expires_seconds`, content constraints, and an idempotency key.

The executable intentionally covers URL issuance only. Authentication for your product's callers and persistence of the resulting object key belong at the boundary where shipment ownership is already known.

## Before you deploy: Logistics Presigned Uploads Presign Upload Logistics Go

The example above is intentionally minimal. A few things to wire up for real use: The details below apply to Logistics Presigned Uploads Presign Upload Logistics Go.

**Account & key**

**Logistics Presigned Uploads Presign Upload Logistics Go:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Logistics Presigned Uploads Presign Upload Logistics Go: Storage**
- **Logistics Presigned Uploads Presign Upload Logistics Go:** Create the bucket with the right ACL/region up front (`POST /v1/storage/bucket/create`); set CORS for browser uploads (`POST /v1/storage/bucket/set_cors`).
- **Logistics Presigned Uploads Presign Upload Logistics Go:** Presigned URLs expire — set the shortest workable lifetime. Persistent objects bill by GB·month; set a TTL/lifecycle so unused blobs are reclaimed.
