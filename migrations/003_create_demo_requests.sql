-- Create demo_requests table for demo request form data
CREATE TABLE IF NOT EXISTS demo_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT NOT NULL,
    company TEXT NOT NULL,
    build_system TEXT NOT NULL CHECK (build_system IN ('yocto', 'buildroot', 'debian', 'custom')),
    message TEXT,
    status TEXT NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'contacted', 'converted')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add index on email for lookups
CREATE INDEX IF NOT EXISTS idx_demo_requests_email ON demo_requests(email);

-- Add index on created_at for sorting
CREATE INDEX IF NOT EXISTS idx_demo_requests_created_at ON demo_requests(created_at DESC);

-- Add index on status for filtering
CREATE INDEX IF NOT EXISTS idx_demo_requests_status ON demo_requests(status);

-- Add index on build_system for analytics
CREATE INDEX IF NOT EXISTS idx_demo_requests_build_system ON demo_requests(build_system);
