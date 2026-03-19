# Mailgun Email Setup

## Why Mailgun?
- **Bypasses Docker/firewall blocks**: Uses HTTPS (port 443) instead of SMTP ports
- **Reliable**: Better deliverability than direct SMTP
- **Free tier**: 10,000 emails/month for 3 months, then pay-as-you-go
- **Simple**: Just HTTP API calls

## Setup Steps

### 1. Sign up for Mailgun
1. Go to https://signup.mailgun.com/new/signup
2. Create account (free trial available)
3. Verify your email

### 2. Get Your Credentials
From your Mailgun dashboard:

**For Sandbox Domain (testing):**
- Domain: `sandboxXXXXXXX.mailgun.org` (shown on dashboard)
- API Key: Click "API Keys" → Copy "Private API key"
- Region: US (or EU if you selected Europe)

**For Production (verify your own domain):**
- Add your domain: `leonacompliance.be`
- Follow DNS setup instructions (add TXT and CNAME records)
- Use verified domain instead of sandbox

### 3. Configure Environment Variables in Coolify

Add these to your app's Environment Variables:

```bash
MAILGUN_API_KEY=your-api-key-here
MAILGUN_DOMAIN=sandboxXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX.mailgun.org
MAILGUN_FROM=LEONA Compliance <support@leonacompliance.be>
MAILGUN_REGION=US
```

**Important:**
- Replace with YOUR actual API key and domain from Mailgun dashboard
- Set `MAILGUN_REGION=EU` if you're using EU servers
- For sandbox, `MAILGUN_FROM` must match an authorized recipient

### 4. Redeploy in Coolify

After adding the environment variables:
1. Save the changes
2. Trigger a new deployment
3. Check logs - you should see:
   ```
   📧 Creating Mailgun email service...
   ✅ Mailgun service initialized (domain: sandboxXXX.mailgun.org, region: US)
   ✅ Mailgun service ready
   ```

### 5. Test

1. Visit your site: `leonacompliance.be`
2. Download the TCF sample report (enter email)
3. Check Mailgun dashboard → Logs to see if email was sent
4. Check your inbox

## Troubleshooting

### "Mailgun not configured"
- Check environment variables are set correctly in Coolify
- Redeploy after adding env vars

### Sandbox limitations
- Can only send to authorized recipients
- Add your email in Mailgun dashboard → Sending → Authorized Recipients
- Or verify your domain for production use

### Still not working?
- Check Coolify logs for errors
- Check Mailgun dashboard → Logs for delivery status
- Verify API key is correct (no extra spaces)

## Production Checklist

Before going live:
- [ ] Verify your domain `leonacompliance.be` in Mailgun
- [ ] Add DNS records (SPF, DKIM, CNAME)
- [ ] Update `MAILGUN_DOMAIN` to your verified domain
- [ ] Remove sandbox domain
- [ ] Test from production
- [ ] Monitor Mailgun analytics dashboard

## Fallback to SMTP

The code automatically falls back to SMTP if Mailgun is not configured. To use SMTP instead:
- Don't set `MAILGUN_API_KEY` or `MAILGUN_DOMAIN`
- Set `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`

## Cost

- **Free tier**: 10,000 emails/month for 3 months
- **After trial**: $35/month for 50,000 emails
- **Pay-as-you-go**: $0.80 per 1,000 emails

For a startup, the free tier is more than enough!
