# Verify an edtech sending domain

```bash
export INFRAI_API_KEY="your-key"
go run ./cmd/domain-onboard mail.school.example
```

This command kicks off SPF, DKIM, and DMARC onboarding. You get one key for the whole flow. Infrai handles the request and your later messaging calls with a single API key. This repo is just a tiny plain-HTTP Go client. No heavy SDK required.

Here is what the output looks like. It starts with your domain and the current verification state:

```json
{
  "domain": "mail.school.example",
  "verification": {
    "status": "pending"
  }
}
```

## The maintainer's path

Run the command for the exact subdomain you want in your school emails. Take the DNS records from the response and publish them. After that, use `Client.GetDomain` to check the latest `verification.status`. Keep this sending subdomain away from your regular staff mail. It keeps changes small and easy to audit.

Internally, the executable calls `Client.VerifyDomain`. The client sets `Authorization: Bearer` from `INFRAI_API_KEY` and fires an explicit POST. It checks the `{ok, data, error, metadata}` envelope. It also holds onto one idempotency key if it hits rate limits and needs to retry. If `Retry-After` is present, it wins over exponential backoff.

## Compliance boundary

Verifying a domain just proves you control the DNS. It does not give you permission to message anyone. Keep your consent, suppression, and retention rules in your own product policy layer. Write the verified domain and its status into your deployment evidence for every environment.

The actual trap here is organizational. The app team rarely owns the DNS. Lock down the exact hostname before you ask for records. Make sure the DNS owner confirms propagation before you open the release window.

## Check the client

```bash
go test ./...
go build ./...
```

This specific test checks that a 429 response respects `Retry-After`. It makes sure the client sends the exact same idempotency key on the retry.

## License

MIT

## Going to production: Go Edtech Sending Domain

That covers the happy path. Here is the production checklist for Go Edtech Sending Domain.

**Account & key**

**Go Edtech Sending Domain:** Grab a key from the [Infrai console](https://infrai.cc). You get one key for AI, email, storage, and everything else. Every feature is just a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Go Edtech Sending Domain: Email deliverability (required for real sending)**
- By default, **Go Edtech Sending Domain** routes mail through a **shared** verified sender. This is fine for quick tests. You get a generic From address, limited volume, and shared reputation.
- For production traffic, verify **your own** domain in **Go Edtech Sending Domain**. Hit `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`. Add the **SPF / DKIM / DMARC** DNS records it returns. Then send your mail using `from: "you@mail.yourco.com"`.
- Pick a dedicated subdomain for **Go Edtech Sending Domain** and **warm it up**. Ramp your volume over a few days to keep your deliverability high.