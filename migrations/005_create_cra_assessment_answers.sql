-- Create table for temporary CRA assessment answers (before final submission)
CREATE TABLE IF NOT EXISTS cra_assessment_answers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    question_number INTEGER NOT NULL,
    answer TEXT NOT NULL CHECK(answer IN ('ja', 'nee')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(session_id, question_number)
);

CREATE INDEX idx_cra_assessment_session ON cra_assessment_answers(session_id);
CREATE INDEX idx_cra_assessment_created ON cra_assessment_answers(created_at);
