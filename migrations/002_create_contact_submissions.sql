-- Create contact_submissions table for contact form data
CREATE TABLE IF NOT EXISTS contact_submissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT NOT NULL,
    company TEXT NOT NULL,
    message TEXT NOT NULL,
    solution TEXT NOT NULL CHECK (solution IN ('snapshot', 'shield', 'pipeline')),
    status TEXT NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'contacted', 'converted')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add index on email for lookups
CREATE INDEX IF NOT EXISTS idx_contact_submissions_email ON contact_submissions(email);

-- Add index on created_at for sorting
CREATE INDEX IF NOT EXISTS idx_contact_submissions_created_at ON contact_submissions(created_at DESC);

-- Add index on status for filtering
CREATE INDEX IF NOT EXISTS idx_contact_submissions_status ON contact_submissions(status);
