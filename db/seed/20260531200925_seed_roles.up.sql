-- SQL Seed up-migration: seed_roles
INSERT INTO roles (id, name, role_type, permissions)
VALUES
('11b38d48-8605-4e1f-8630-2c2120fbd682', 'SUPER_ADMIN', null, '{"FULLACCESS": true}'),
('d4df0794-22e8-4a30-9039-bbcd76447b56', 'SUB_ADMIN', null, '{"product": ["read:product", "write:product"]}');

