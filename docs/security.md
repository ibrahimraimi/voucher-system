# Security Model

The system employs a defense-in-depth strategy to secure voucher codes.

## 1. PIN Structure
A 16-digit numeric PIN is composed of:
- **Entropy**: 11 random digits (cryptographically secure).
- **Integrity**: 1 Luhn checksum digit (detects typos).
- **Authenticity**: 4 HMAC signature digits.

`[ 11 Random Digits ] [ 1 Luhn Digit ] [ 4 HMAC Digits ]`

This structure ensures that valid PINs cannot be easily guessed, and random strings will be rejected before touching the database.

## 2. Storage Security
- **Hashing**: PINs are **never** stored in plaintext. They are hashed using `SHA-256`.
- **Why SHA-256?**: Fast and secure for this use case. Use of bcrypt/argon2 is deliberately avoided for high-throughput redemption scenarios, relying instead on the large search space (10^16) and rate limiting to prevent brute forcing.
- **Timing Attack Prevention**: All hash comparisons use `crypto/subtle.ConstantTimeCompare`.

## 3. Runtime Protection
- **Rate Limiting**: Users are identified by ID (or IP in a web context).
  - Limit: Configurable (default 3 attempts).
  - Lockout: Temporary block (e.g., 1 minute) after threshold.
- **Double Spending**: Strict `status` check within an ACID transaction ensures a voucher cannot be redeemed twice, even under heavy concurrency.

## 4. Threat Model & Mitigations

| Threat | Mitigation |
|--------|------------|
| **Brute Force Guessing** | HMAC signature makes guessing 1/10000 *per attempt*, plus Rate Limiting locks attackers out. |
| **SQL Injection** | Parameterized queries used everywhere. |
| **Database Leak** | Only hashes are stored; attacker cannot reverse hashes to get usable PINs easily without significant compute (and even then, search space is large). |
| **Race Conditions** | Database-level atomic updates (`UPDATE ... WHERE status='unused'`). |
