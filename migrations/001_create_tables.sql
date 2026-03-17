-- LEONA Database Schema
-- Supabase PostgreSQL Migration

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Scans Table: Stores all SBOM uploads and analysis results
CREATE TABLE scans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, analyzing, completed, failed
    payment_status VARCHAR(50) NOT NULL DEFAULT 'free', -- free, pending_payment, paid
    compliance_score INTEGER,
    sbom_format VARCHAR(50), -- cyclonedx, spdx, custom
    sbom_size_kb INTEGER,
    total_components INTEGER,
    kernel_version VARCHAR(100),
    critical_findings INTEGER DEFAULT 0,
    hostname VARCHAR(255),
    company_name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    analyzed_at TIMESTAMP WITH TIME ZONE,
    paid_at TIMESTAMP WITH TIME ZONE
);

-- Payments Table: Tracks Mollie payments
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scan_id UUID REFERENCES scans(id) ON DELETE CASCADE,
    mollie_payment_id VARCHAR(255) UNIQUE NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'EUR',
    status VARCHAR(50) NOT NULL, -- open, paid, failed, canceled, expired
    payment_method VARCHAR(50), -- bancontact, ideal, creditcard
    tier VARCHAR(20) NOT NULL, -- tier1, tier2, tier3
    customer_email VARCHAR(255) NOT NULL,
    customer_name VARCHAR(255),
    description TEXT,
    redirect_url TEXT,
    webhook_url TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    paid_at TIMESTAMP WITH TIME ZONE,
    canceled_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE
);

-- Leads Table: Marketing lead capture (engineers + lawyers)
CREATE TABLE leads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    lead_type VARCHAR(20) NOT NULL, -- engineer, lawyer, general
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    company_name VARCHAR(255),
    job_title VARCHAR(100),
    phone VARCHAR(50),
    source VARCHAR(50) DEFAULT 'website', -- website, linkedin, referral
    status VARCHAR(50) DEFAULT 'new', -- new, contacted, qualified, converted
    notes TEXT,
    lead_magnet_requested VARCHAR(100), -- bitbake_layer, annex_template, free_report
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_contacted_at TIMESTAMP WITH TIME ZONE,
    converted_at TIMESTAMP WITH TIME ZONE
);

-- Analysis Results Table: Detailed technical findings
CREATE TABLE analysis_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scan_id UUID REFERENCES scans(id) ON DELETE CASCADE,
    framework VARCHAR(20) NOT NULL, -- CRA, CER, NIS2
    requirement_id VARCHAR(50) NOT NULL,
    requirement_description TEXT,
    status VARCHAR(50) NOT NULL, -- compliant, partial, non_compliant
    finding TEXT,
    remediation TEXT,
    severity VARCHAR(20), -- critical, high, medium, low
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Downloads Table: Track PDF and fix-pack downloads
CREATE TABLE downloads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scan_id UUID REFERENCES scans(id) ON DELETE CASCADE,
    download_type VARCHAR(50) NOT NULL, -- pdf, fixes, attestation_script
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_scans_email ON scans(email);
CREATE INDEX idx_scans_status ON scans(status);
CREATE INDEX idx_scans_payment_status ON scans(payment_status);
CREATE INDEX idx_scans_created_at ON scans(created_at DESC);
CREATE INDEX idx_payments_mollie_id ON payments(mollie_payment_id);
CREATE INDEX idx_payments_scan_id ON payments(scan_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_leads_email ON leads(email);
CREATE INDEX idx_leads_lead_type ON leads(lead_type);
CREATE INDEX idx_leads_status ON leads(status);
CREATE INDEX idx_analysis_results_scan_id ON analysis_results(scan_id);
CREATE INDEX idx_downloads_scan_id ON downloads(scan_id);

-- Row Level Security (RLS)
ALTER TABLE scans ENABLE ROW LEVEL SECURITY;
ALTER TABLE payments ENABLE ROW LEVEL SECURITY;
ALTER TABLE leads ENABLE ROW LEVEL SECURITY;
ALTER TABLE analysis_results ENABLE ROW LEVEL SECURITY;
ALTER TABLE downloads ENABLE ROW LEVEL SECURITY;

-- Policies: Allow service role full access
CREATE POLICY "Service role can do everything on scans" ON scans
    FOR ALL USING (auth.role() = 'service_role');

CREATE POLICY "Service role can do everything on payments" ON payments
    FOR ALL USING (auth.role() = 'service_role');

CREATE POLICY "Service role can do everything on leads" ON leads
    FOR ALL USING (auth.role() = 'service_role');

CREATE POLICY "Service role can do everything on analysis_results" ON analysis_results
    FOR ALL USING (auth.role() = 'service_role');

CREATE POLICY "Service role can do everything on downloads" ON downloads
    FOR ALL USING (auth.role() = 'service_role');

-- Functions
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers
CREATE TRIGGER update_scans_updated_at BEFORE UPDATE ON scans
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_leads_updated_at BEFORE UPDATE ON leads
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Views for analytics
CREATE VIEW scan_analytics AS
SELECT 
    DATE_TRUNC('day', created_at) as date,
    COUNT(*) as total_scans,
    COUNT(*) FILTER (WHERE payment_status = 'paid') as paid_scans,
    COUNT(*) FILTER (WHERE payment_status = 'free') as free_scans,
    AVG(compliance_score) as avg_compliance_score,
    COUNT(*) FILTER (WHERE critical_findings > 0) as scans_with_critical_findings
FROM scans
GROUP BY DATE_TRUNC('day', created_at)
ORDER BY date DESC;

CREATE VIEW revenue_analytics AS
SELECT 
    DATE_TRUNC('day', paid_at) as date,
    COUNT(*) as total_payments,
    SUM(amount) as total_revenue,
    AVG(amount) as avg_transaction_value,
    COUNT(*) FILTER (WHERE tier = 'tier1') as tier1_sales,
    COUNT(*) FILTER (WHERE tier = 'tier2') as tier2_sales,
    COUNT(*) FILTER (WHERE tier = 'tier3') as tier3_sales
FROM payments
WHERE status = 'paid'
GROUP BY DATE_TRUNC('day', paid_at)
ORDER BY date DESC;

-- Comments
COMMENT ON TABLE scans IS 'Stores all SBOM uploads and analysis results';
COMMENT ON TABLE payments IS 'Tracks all Mollie payment transactions';
COMMENT ON TABLE leads IS 'Marketing lead capture from website forms';
COMMENT ON TABLE analysis_results IS 'Detailed compliance findings per framework';
COMMENT ON TABLE downloads IS 'Tracks PDF and fix-pack downloads for conversion analysis';
