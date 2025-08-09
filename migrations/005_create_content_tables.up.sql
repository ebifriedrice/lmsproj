-- Video content table
CREATE TABLE videos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    lesson_id INTEGER NOT NULL UNIQUE, -- Enforces one video per lesson
    title TEXT NOT NULL,
    video_url TEXT NOT NULL,
    FOREIGN KEY (lesson_id) REFERENCES lessons(id) ON DELETE CASCADE
);

-- Text content table
CREATE TABLE texts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    lesson_id INTEGER NOT NULL UNIQUE, -- Enforces one text content per lesson
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    FOREIGN KEY (lesson_id) REFERENCES lessons(id) ON DELETE CASCADE
);

-- Multiple Choice Question content table
CREATE TABLE mcqs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    lesson_id INTEGER NOT NULL UNIQUE, -- Enforces one MCQ per lesson
    question TEXT NOT NULL,
    -- Storing options as a JSON array of strings
    options TEXT NOT NULL,
    -- Storing the index of the correct option in the JSON array
    correct_option_index INTEGER NOT NULL,
    FOREIGN KEY (lesson_id) REFERENCES lessons(id) ON DELETE CASCADE
);

-- Table to store student submissions/scores for MCQs
CREATE TABLE mcq_submissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    mcq_id INTEGER NOT NULL,
    selected_option_index INTEGER NOT NULL,
    is_correct BOOLEAN NOT NULL,
    submitted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (mcq_id) REFERENCES mcqs(id) ON DELETE CASCADE,
    -- A user can only submit an answer for an MCQ once
    UNIQUE (user_id, mcq_id)
);
