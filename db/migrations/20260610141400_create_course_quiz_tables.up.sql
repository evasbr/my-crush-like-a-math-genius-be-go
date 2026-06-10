CREATE TABLE topics (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    level_settings JSONB,
    max_attempts INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP
);

CREATE TABLE questions (
    id UUID PRIMARY KEY,
    topic_id UUID NOT NULL,
    content TEXT NOT NULL,
    level VARCHAR(50) NOT NULL,
    time_limit INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP,
    CONSTRAINT fk_questions_topic FOREIGN KEY (topic_id) REFERENCES topics (id) ON DELETE CASCADE
);

CREATE TABLE answers (
    id UUID PRIMARY KEY,
    question_id UUID NOT NULL,
    content TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_answers_question FOREIGN KEY (question_id) REFERENCES questions (id) ON DELETE CASCADE
);

CREATE TABLE attempt_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    topic_id UUID NOT NULL,
    selected_level VARCHAR(50) NOT NULL,
    requested_questions INT NOT NULL,
    score INT,
    status VARCHAR(50) NOT NULL,
    started_at TIMESTAMP NOT NULL DEFAULT now(),
    finished_at TIMESTAMP,
    meta_data JSONB,
    CONSTRAINT fk_attempt_sessions_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_attempt_sessions_topic FOREIGN KEY (topic_id) REFERENCES topics (id) ON DELETE CASCADE
);

CREATE TABLE attempt_details (
    id UUID PRIMARY KEY,
    attempt_session_id UUID NOT NULL,
    question_id UUID NOT NULL,
    answer_id UUID,
    is_correct BOOLEAN,
    answered_at TIMESTAMP,
    CONSTRAINT fk_attempt_details_session FOREIGN KEY (attempt_session_id) REFERENCES attempt_sessions (id) ON DELETE CASCADE,
    CONSTRAINT fk_attempt_details_question FOREIGN KEY (question_id) REFERENCES questions (id) ON DELETE CASCADE,
    CONSTRAINT fk_attempt_details_answer FOREIGN KEY (answer_id) REFERENCES answers (id) ON DELETE SET NULL
);
