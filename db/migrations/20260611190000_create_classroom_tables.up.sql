CREATE TABLE classrooms (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    codes VARCHAR(50) UNIQUE NOT NULL,
    cover_img TEXT,
    wallpaper_img TEXT,
    is_external_invite_enable BOOLEAN NOT NULL DEFAULT TRUE,
    invite_link TEXT,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP
);

CREATE TABLE classroom_roles (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    classroom_id UUID NOT NULL,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP,
    CONSTRAINT fk_classroom_roles_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_classroom_roles_classroom FOREIGN KEY (classroom_id) REFERENCES classrooms(id) ON DELETE CASCADE,
    CONSTRAINT chk_classroom_roles_role CHECK (role IN ('owner', 'teacher', 'student'))
);

CREATE UNIQUE INDEX idx_classroom_roles_user_classroom ON classroom_roles(user_id, classroom_id) WHERE deleted_at IS NULL;

-- Truncate existing quiz tables to make sure NOT NULL classroom_id constraint can be successfully added to topics
TRUNCATE TABLE topics, questions, answers, attempt_sessions, attempt_details CASCADE;

ALTER TABLE topics ADD COLUMN classroom_id UUID NOT NULL;
ALTER TABLE topics ADD CONSTRAINT fk_topics_classroom FOREIGN KEY (classroom_id) REFERENCES classrooms(id) ON DELETE CASCADE;
ALTER TABLE topics ADD COLUMN description TEXT;

ALTER TABLE topics ADD COLUMN female_normal_img TEXT NOT NULL;
ALTER TABLE topics ADD COLUMN male_normal_img TEXT NOT NULL;

ALTER TABLE topics ADD COLUMN female_dating_img TEXT NOT NULL;
ALTER TABLE topics ADD COLUMN male_dating_img TEXT NOT NULL;

ALTER TABLE topics ADD COLUMN female_normal_dialog TEXT NOT NULL;
ALTER TABLE topics ADD COLUMN male_normal_dialog TEXT NOT NULL;

ALTER TABLE topics ADD COLUMN female_dating_dialog TEXT NOT NULL;
ALTER TABLE topics ADD COLUMN male_dating_dialog TEXT NOT NULL;

ALTER TABLE topics ADD COLUMN status VARCHAR(50) NOT NULL;
