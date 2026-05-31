INSERT INTO tb_user (username, password, is_active) VALUES 
('admin', '$2a$10$tZ8NlP0yS.y6KxY.PZ6xreK3q3aIebR6kUfP41sQ1XyB9aV7XvN.a', true)
ON CONFLICT (username) DO NOTHING;

INSERT INTO tb_user_role (user_role_id, role, username) VALUES 
('e0d4c9a1-0f72-4d0f-90db-3b56cfc4ad72', 'ROLE_ADMIN', 'admin')
ON CONFLICT (user_role_id) DO NOTHING;
