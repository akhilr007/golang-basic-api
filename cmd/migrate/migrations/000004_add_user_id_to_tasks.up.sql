ALTER TABLE tasks ADD COLUMN user_id INT;

UPDATE tasks SET user_id = 1 WHERE user_id IS NULL;

ALTER TABLE tasks ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE tasks
ADD CONSTRAINT fk_tasks_user
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

CREATE INDEX idx_tasks_user_id ON tasks(user_id);
