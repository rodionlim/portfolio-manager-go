# Dividends.sg duplicate distributions and recovery

Dividends.sg can expose legacy corporate-action rows alongside refreshed rows
carrying an SGX reference. Adding every row by ex-date inflated some historical
dividend totals in September 2026.

The parser matches legacy and referenced rows one-for-one only when ex-date,
payment date, currency, amount, and normalized particulars agree. Separate SGX
references and unmatched legacy events remain distinct, even if their amounts
are equal. Matching does not discard separate taxable, tax-exempt, or capital
components. This is a conservative compatibility rule for the observed upstream
duplication, not a general corporate-action revision resolver.

Bond coupon parsing excludes source annotations and rejects principal-sized
percentage rates (100% or more) rather than interpreting them as annual coupons.
The existing semiannual coupon assumption remains; redemption proceeds are not
calculated as dividends by this parser.

## Recovery after deploying the fix

1. Restart the backend with the corrected parser. This clears both the source
   dividend cache and the dividends manager's in-memory cache (normally 24 hours).
2. Request positions and portfolio metrics normally. On successful fetches, the
   source updates stored official dividend totals for the returned ex-dates,
   including replacing an inflated total with a smaller corrected total.
3. Compare current dividends and P&L with the reconciliation before deciding
   whether any further data repair is necessary.

No database wipe, schema migration, or custom-dividend override is required for
affected official dates still returned by the source. Older dates no longer
visible upstream are retained, and custom overrides retain precedence. An
affected date hidden upstream, a custom override containing a wrong amount, or
an unsuccessful fetch will not automatically be corrected by this process.

Historical metrics snapshots are separate, stored aggregate values. Refreshing
dividend metadata does not rewrite an already inflated snapshot. Review affected
snapshots separately; do not substitute today's valuation for a past snapshot.

This fix does not change the summary's current-FX cost conversion or the separate
cross-book dividend attribution issue identified during the reconciliation.
