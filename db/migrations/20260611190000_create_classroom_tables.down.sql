ALTER TABLE topics DROP CONSTRAINT IF EXISTS fk_topics_classroom;
ALTER TABLE topics DROP COLUMN IF EXISTS classroom_id;
ALTER TABLE topics DROP COLUMN IF EXISTS description;
ALTER TABLE topics DROP COLUMN IF EXISTS female_normal_img;
ALTER TABLE topics DROP COLUMN IF EXISTS male_normal_img;
ALTER TABLE topics DROP COLUMN IF EXISTS female_dating_img;
ALTER TABLE topics DROP COLUMN IF EXISTS male_dating_img;
ALTER TABLE topics DROP COLUMN IF EXISTS female_normal_dialog;
ALTER TABLE topics DROP COLUMN IF EXISTS male_normal_dialog;
ALTER TABLE topics DROP COLUMN IF EXISTS female_dating_dialog;
ALTER TABLE topics DROP COLUMN IF EXISTS male_dating_dialog;
ALTER TABLE topics DROP COLUMN IF EXISTS status;

DROP TABLE IF EXISTS classroom_roles;
DROP TABLE IF EXISTS classrooms;
