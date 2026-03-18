-- Create snapshot_submissions table for detailed snapshot audit requests
CREATE TABLE IF NOT EXISTS snapshot_submissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Order tracking
    order_uuid TEXT NOT NULL UNIQUE,
    
    -- Contact information
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT NOT NULL,
    company TEXT NOT NULL,
    phone TEXT,
    
    -- Build system details
    build_system TEXT NOT NULL,
    build_system_version TEXT,
    target_architecture TEXT NOT NULL,
    kernel_version TEXT NOT NULL,
    libc TEXT NOT NULL,
    
    -- Product information
    product_name TEXT NOT NULL,
    product_category TEXT NOT NULL,
    connectivity JSONB, -- Array of connectivity options
    annual_volume TEXT NOT NULL,
    
    -- Security features
    secure_boot TEXT,
    tpm TEXT,
    ota_features JSONB, -- Array of OTA features
    update_framework TEXT,
    
    -- Artifacts transfer
    artifact_access TEXT NOT NULL,
    estimated_size TEXT,
    available_artifacts JSONB, -- Array of available artifacts
    
    -- Additional context
    timeline TEXT,
    concerns TEXT,
    additional_notes TEXT,
    
    -- Legal
    nda_accepted BOOLEAN NOT NULL DEFAULT false,
    
    -- Payment tracking
    payment_status TEXT NOT NULL DEFAULT 'pending' CHECK (payment_status IN ('pending', 'paid', 'failed', 'refunded')),
    mollie_payment_id TEXT,
    payment_completed_at TIMESTAMPTZ,
    
    -- Status tracking
    status TEXT NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'payment-pending', 'paid', 'in-progress', 'completed', 'cancelled')),
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add indexes for common queries
CREATE INDEX IF NOT EXISTS idx_snapshot_order_uuid ON snapshot_submissions(order_uuid);
CREATE INDEX IF NOT EXISTS idx_snapshot_email ON snapshot_submissions(email);
CREATE INDEX IF NOT EXISTS idx_snapshot_company ON snapshot_submissions(company);
CREATE INDEX IF NOT EXISTS idx_snapshot_payment_status ON snapshot_submissions(payment_status);
CREATE INDEX IF NOT EXISTS idx_snapshot_status ON snapshot_submissions(status);
CREATE INDEX IF NOT EXISTS idx_snapshot_created_at ON snapshot_submissions(created_at DESC);

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_snapshot_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_snapshot_updated_at
    BEFORE UPDATE ON snapshot_submissions
    FOR EACH ROW
    EXECUTE FUNCTION update_snapshot_updated_at();
