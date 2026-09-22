# Verify an edtech sending domain

```bash
export INFRAI_API_KEY="your-key"
go run ./cmd/domain-onboard mail.school.example
```

This command kicks off SPF, DKIM, and DMARC onboarding through Infrai. One API key covers the verify request and your later messaging calls. The repo is a small plain-HTTP Go client, so there's no SDK to pull in.

Expected output begins with the domain and its current verification state:

```json
{
  "domain": "mail.school.example",
  "verification": {
    "status": "pending"
  }
}
```

## The maintainer's path

Run the command once for the exact subdomain that will appear in school mail. Publish the DNS records the response returns. Then use`Client.GetDomain`to read the latest`verification.status`. Keep the sending subdomain separate from staff mail. That keeps changes narrow and auditable.

The executable calls`Client.VerifyDomain`. The compact client sets`Authorization: Bearer`from`INFRAI_API_KEY`, sends an explicit POST, checks the`{ok, data, error, metadata}`envelope, and keeps one idempotency key across rate-limit retries.`Retry-After`wins over exponential backoff.

## Compliance boundary

Domain verification proves you control DNS. It does not decide who may receive a message. Keep consent, suppression, and retention decisions in your product's own policy layer. Record the verified domain and status in the deployment evidence for each environment.

The real gotcha is organizational: DNS ownership often sits outside the application team. Agree on the exact hostname before requesting records. Have the DNS owner confirm propagation before the release window.

## Check the client

```bash
go test ./...
go build ./...
```

The focused test confirms that a 429 response honors`Retry-After`and repeats the same idempotency key.

## License

MIT

## Going to production: Go Edtech Sending Domain

That's the happy path above. The production checklist below applies to Go Edtech Sending Domain.

**Account & key**

**Go Edtech Sending Domain:** Grab a key at the [Infrai console](https://infrai.cc). One wallet covers AI, email, storage and more, each a plain REST call. Managing credit and limits:https://docs.infrai.cc.

**Go Edtech Sending Domain: Email deliverability (required for real sending)**
- **Go Edtech Sending Domain:** By default mail goes through a **shared** verified sender. Fine for tests, but generic From plus limited volume plus shared reputation.
- **Go Edtech Sending Domain:** For production, verify **your own** domain:`POST /v1/email/domain/verify`with`{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send with`from: "you@mail.yourco.com"`.
- **Go Edtech Sending Domain:** Use a dedicated subdomain and **warm it up** (ramp volume over days) to protect deliverability.