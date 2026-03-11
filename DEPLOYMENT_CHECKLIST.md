# LEONA Scanner - Production Deployment Checklist

## Pre-Deployment (15 minutes)

### 1. Supabase Setup
- [ ] Create Supabase project at https://supabase.com
- [ ] Copy Project URL from Project Settings > API
- [ ] Copy `service_role` key (NOT anon key) from Project Settings > API
- [ ] Open SQL Editor in Supabase dashboard
- [ ] Copy contents of `migrations/001_create_tables.sql`
- [ ] Paste and run in SQL Editor
- [ ] Verify tables created: scans, payments, leads, analysis_results, downloads
- [ ] Check that RLS policies are enabled

### 2. Mollie Setup
- [ ] Create Mollie account at https://www.mollie.com/dashboard/signup
- [ ] Complete KYC verification (business details, bank account)
- [ ] Get test API key from Developers > API Keys (starts with `test_`)
- [ ] Get live API key when ready for production (starts with `live_`)
- [ ] Set webhook URL in Mollie dashboard: `https://yourdomain.com/webhook/mollie`

### 3. Environment Variables

Update your production `.env`:

```bash
# Supabase
SUPABASE_URL=https://xxxxx.supabase.co
SUPABASE_SERVICE_KEY=eyJhbGc...
SUPABASE_ANON_KEY=eyJhbGc...  # Optional, for RLS

# Mollie
MOLLIE_API_KEY=live_xxxxx  # Use test_xxxxx for testing

# Server
PORT=8080
BASE_URL=https://leona-cravit.be
```

### 4. Code Integration

Edit your main server file (e.g., `cmd/server/main.go`):

- [ ] Add Supabase client initialization (see `SUPABASE_INTEGRATION_EXAMPLE.md`)
- [ ] Update upload handler to call `db.CreateScan()` and `db.CreateLead()`
- [ ] Update analysis handler to call `db.UpdateScanStatus()` and `db.SaveAnalysisResults()`
- [ ] Create payment endpoint that uses Mollie API
- [ ] Store scan_id in session/JWT for payment association

### 5. Webhook Server Deployment

Build webhook server:
```bash
cd cmd/webhook
go build -o webhook
```

Create systemd service `/etc/systemd/system/leona-webhook.service`:
```ini
[Unit]
Description=LEONA Mollie Webhook Server
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/leona-scanner
Environment="SUPABASE_URL=https://xxxxx.supabase.co"
Environment="SUPABASE_SERVICE_KEY=eyJhbGc..."
Environment="MOLLIE_API_KEY=live_xxxxx"
Environment="PORT=8081"
ExecStart=/opt/leona-scanner/webhook
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable leona-webhook
sudo systemctl start leona-webhook
sudo systemctl status leona-webhook
```

### 6. Nginx Configuration

Add webhook proxy to your nginx config:

```nginx
server {
    listen 80;
    server_name leona-cravit.be;

    # Main application
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # Webhook endpoint
    location /webhook/mollie {
        proxy_pass http://localhost:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

Reload nginx:
```bash
sudo nginx -t
sudo systemctl reload nginx
```

## Testing (10 minutes)

### Test Free Scan Flow
- [ ] Visit homepage
- [ ] Upload SBOM file with email
- [ ] Check Supabase dashboard: verify scan record created in `scans` table
- [ ] Check Supabase dashboard: verify lead created in `leads` table
- [ ] Verify email received with scan ID
- [ ] Check scan shows 65% compliance score

### Test Payment Flow (Test Mode)
- [ ] Click "Unlock Full Report" on scan result
- [ ] Verify redirected to Mollie checkout
- [ ] Use test card: `4111 1111 1111 1111`, exp: any future date, CVC: 123
- [ ] Complete test payment
- [ ] Verify redirected to success page
- [ ] Check Supabase `payments` table: payment status = "paid"
- [ ] Check Supabase `scans` table: payment_status = "paid"
- [ ] Verify download links appear
- [ ] Check webhook server logs: `sudo journalctl -u leona-webhook -f`

### Test Analytics
Run queries in Supabase SQL Editor:

```sql
-- Today's scans
SELECT COUNT(*) FROM scans WHERE created_at::date = CURRENT_DATE;

-- Conversion rate
SELECT * FROM scan_analytics WHERE date = CURRENT_DATE;

-- Revenue
SELECT SUM(amount) FROM payments WHERE status = 'paid';

-- Hot leads (multiple scans, no payment)
SELECT email, COUNT(*) as scan_count 
FROM scans 
WHERE payment_status = 'free' 
GROUP BY email 
HAVING COUNT(*) > 1;
```

## Go Live (5 minutes)

- [ ] Switch Mollie API key from `test_` to `live_`
- [ ] Update `MOLLIE_API_KEY` in both server and webhook `.env`
- [ ] Restart both services:
  ```bash
  sudo systemctl restart leona-server
  sudo systemctl restart leona-webhook
  ```
- [ ] Test one real payment with your own card (€1 test)
- [ ] Verify payment appears in Mollie dashboard
- [ ] Refund test payment in Mollie dashboard

## Post-Launch Monitoring

### Day 1 Checks
- [ ] Monitor webhook logs: `sudo journalctl -u leona-webhook -f`
- [ ] Check Supabase logs in dashboard (Project Settings > Logs)
- [ ] Verify scan_analytics view populating
- [ ] Test one manual payment flow end-to-end

### Weekly Analytics Queries

**Conversion Funnel:**
```sql
SELECT 
    date,
    total_scans,
    paid_scans,
    ROUND(100.0 * paid_scans / NULLIF(total_scans, 0), 2) as conversion_rate
FROM scan_analytics
WHERE date >= CURRENT_DATE - INTERVAL '7 days'
ORDER BY date DESC;
```

**Revenue by Tier:**
```sql
SELECT 
    tier,
    COUNT(*) as count,
    SUM(amount) as total_revenue,
    AVG(amount) as avg_revenue
FROM payments
WHERE status = 'paid'
  AND created_at >= CURRENT_DATE - INTERVAL '7 days'
GROUP BY tier;
```

**Hot Leads to Follow Up:**
```sql
SELECT 
    l.email,
    l.company_name,
    COUNT(s.id) as scan_count,
    MAX(s.created_at) as last_scan,
    l.lead_type
FROM leads l
JOIN scans s ON s.email = l.email
WHERE s.payment_status = 'free'
  AND l.status = 'new'
GROUP BY l.email, l.company_name, l.lead_type
HAVING COUNT(s.id) >= 2
ORDER BY scan_count DESC;
```

## Troubleshooting

### Webhook not receiving events
```bash
# Check if webhook server is running
sudo systemctl status leona-webhook

# Check logs
sudo journalctl -u leona-webhook -n 50

# Test manually
curl -X POST http://localhost:8081/health

# Verify Mollie can reach your server
# Check firewall: port 8081 should NOT be public (only internal proxy)
# Mollie hits nginx on port 80, nginx proxies to 8081
```

### Payment stuck in "pending"
- Check Mollie dashboard for payment status
- Manually trigger webhook: Dashboard > Payments > Click payment > Resend webhook
- Check webhook server logs for errors
- Verify scan_id in payment metadata is valid UUID

### Database connection errors
```bash
# Test Supabase connection
curl -H "apikey: YOUR_SERVICE_KEY" \
     -H "Authorization: Bearer YOUR_SERVICE_KEY" \
     https://xxxxx.supabase.co/rest/v1/scans?limit=1

# Check RLS policies (should use service_role key to bypass RLS)
```

## Success Metrics

**First Week Goals:**
- [ ] 20+ free scans completed
- [ ] 5+ leads captured (engineers with company emails)
- [ ] 2+ paid conversions (€998+ revenue)
- [ ] 1 lawyer partnership initiated

**Track in Supabase:**
```sql
-- Weekly dashboard
SELECT 
    'Total Scans' as metric,
    COUNT(*) as value
FROM scans
WHERE created_at >= CURRENT_DATE - INTERVAL '7 days'
UNION ALL
SELECT 
    'Paid Scans',
    COUNT(*)
FROM scans
WHERE payment_status = 'paid'
  AND paid_at >= CURRENT_DATE - INTERVAL '7 days'
UNION ALL
SELECT 
    'Revenue',
    SUM(amount)
FROM payments
WHERE status = 'paid'
  AND paid_at >= CURRENT_DATE - INTERVAL '7 days';
```

## Next Steps After Launch

1. **Set up email alerts** for new paid conversions using Supabase Database Webhooks
2. **Export hot leads CSV** weekly and email manually
3. **Create Supabase dashboard** using Metabase or Grafana
4. **Add lawyer master account** login flow (after first partnership)
5. **Implement affiliate tracking** (add `utm_source` to leads table)
