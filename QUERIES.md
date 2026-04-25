## Query Reporting

### Summary Query Agreagation

```sql
select 
  v.code, 
  count(1) total, 
  SUM(
    CASE WHEN c.status = 'REDEEMED' THEN 1 ELSE 0 END
  ) total_redemptions, 
  COUNT(1) total_claims, 
  sum(discount_applied) total_discount_granted, 
  avg(transaction_amount) avg_transaction_amount 
from 
  claims c 
  join vouchers v on v.id = c.voucher_id 
group by 
  v.code;
```

Explaination:
- query is optimize agregation by voucher code indexed sorted, also join was optimized by using index on voucher_id join
- calclucation on the go using CASE WHEN to assert wether to sum/count or not in one run operation agregation


schema for table

```sql
CREATE TABLE IF NOT EXISTS vouchers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(255) UNIQUE NOT NULL,
    discount_type VARCHAR(50) NOT NULL, 
    discount_value NUMERIC(15, 2) NOT NULL,
    max_uses INT NOT NULL,
    total_claims INT DEFAULT 0 NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    min_transaction_amount NUMERIC(15, 2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS claims (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    voucher_id UUID REFERENCES vouchers(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL, 
    transaction_amount NUMERIC(15, 2),
    discount_applied NUMERIC(15, 2),
    final_amount NUMERIC(15, 2),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(voucher_id, user_id)
);

CREATE INDEX idx_vouchers_code ON vouchers(code);
CREATE INDEX idx_claims_status_voucher ON claims(status, voucher_id);
CREATE INDEX idx_claims_user_voucher ON claims(user_id, voucher_id);
```

- added index to claims.voucher_id and vouchers.id to optimize join and grouping agregate because sorted
- added index to claims.status and claims.voucher_id to optimize filter
- added index to claims.user_id and claims.voucher_id to optimize filter